package export

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"main/internal/model"
	"main/internal/repository"
	"strings"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ExportService struct {
	DownloadsRepo *repository.DownloadListRepo
	Db            *gorm.DB
	TotalNum      int `json:"total_num"`
	SheetName     string
	nextRow       int // 记录下一次要写入的行号
}

func NewExportService(Db *gorm.DB, DownloadsRepo *repository.DownloadListRepo) *ExportService {
	return &ExportService{
		DownloadsRepo: DownloadsRepo,
		Db:            Db,
		SheetName:     "sheet1",
	}
}

func (r *ExportService) ExportData(ctx context.Context, baseSql string, filename string) (err error) {
	err = r.getDataTotalNum(baseSql)
	if err != nil {
		return err
	}
	if r.TotalNum < 1 {
		log.Fatalf("暂无数据: %v", err)
		return
	}
	batchSize := 500
	totalBatches := (r.TotalNum + batchSize - 1) / batchSize
	var globalHeader []string
	var excelFile *excelize.File
	var sw *excelize.StreamWriter
	task := &model.DownloadList{Status: 2}
	go func() {
		r.DownloadsRepo.UpdateTask(filename, task)
	}()
	defer func() {
		if err := excelFile.Close(); err != nil {
			log.Fatalf("关闭 Excel 文件失败: %v", err)
		}
	}()
	for batchNum := 1; batchNum <= totalBatches; batchNum++ {
		data, currentHeader, err := r.findData(baseSql, batchNum, batchSize)
		if err != nil {
			log.Fatalf("第 %d 批查询失败: %v", batchNum, err)
		}
		if len(data) == 0 {
			log.Printf("第 %d 批无数据，跳过", batchNum)
			continue
		}
		if excelFile == nil {
			globalHeader = currentHeader
			excelFile, sw = r.createExcel(globalHeader, data)
			log.Printf("第 %d 批：初始化 Excel 并写入第一批数据", batchNum)
		} else {
			r.appendDataToExcel(data, sw)
			log.Printf("第 %d 批：追加数据完成", batchNum)
		}
	}
	if excelFile != nil {
		if err := sw.Flush(); err != nil {
			panic(err)
		}
		savePath := fmt.Sprintf("static/%s.xlsx", filename)
		if err := excelFile.SaveAs(savePath); err != nil {
			log.Fatalf("保存 Excel 文件失败: %v", err)
		}
		task.Status = 3
		task.Path = savePath
		task.Name = filename
		r.DownloadsRepo.UpdateTask(filename, task)
		log.Printf("所有数据处理完成，Excel 文件已保存至：%s", savePath)
	} else {
		fmt.Println("无有效数据，未生成 Excel 文件")
	}
	return
}
func (r *ExportService) findData(sqlString string, batchNum int, batchSize int) ([][]interface{}, []string, error) {
	offset := (batchNum - 1) * batchSize
	pagedSQL := fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlString, batchSize, offset)

	rows, err := r.Db.Raw(pagedSQL).Rows()
	if err != nil {
		return nil, nil, fmt.Errorf("执行 SQL 失败: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("获取字段名失败: %v", err)
	}
	if len(columns) == 0 {
		return nil, nil, errors.New("查询结果无字段")
	}

	var data [][]interface{}
	values := make([]interface{}, len(columns))
	valuePointers := make([]interface{}, len(columns))
	for i := range values {
		valuePointers[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePointers...); err != nil {
			return nil, nil, fmt.Errorf("扫描行数据失败: %v", err)
		}
		row := make([]interface{}, len(columns))
		for i, val := range values {
			switch v := val.(type) {
			case sql.NullString:
				if v.Valid {
					row[i] = v.String
				} else {
					row[i] = ""
				}
			case sql.NullInt64:
				if v.Valid {
					row[i] = v.Int64
				} else {
					row[i] = 0
				}
			case sql.NullFloat64:
				if v.Valid {
					row[i] = v.Float64
				} else {
					row[i] = 0.0
				}
			case sql.NullTime:
				if v.Valid {
					row[i] = v.Time.Format("2006-01-02 15:04:05")
				} else {
					row[i] = ""
				}
			default:
				row[i] = val
			}
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("遍历结果失败: %v", err)
	}
	return data, columns, nil
}
func (r *ExportService) createExcel(headers []string, data [][]interface{}) (*excelize.File, *excelize.StreamWriter) {
	f := excelize.NewFile()
	sw, err := f.NewStreamWriter(r.SheetName)
	if err != nil {
		log.Fatalf("创建自定义工作表失败: %v", err)
	}
	headerCell, err := excelize.CoordinatesToCellName(1, 1)
	if err != nil {
		log.Fatalf("生成表头单元格名称失败: %v", err)
	}
	headerData := make([]interface{}, len(headers))
	for i, h := range headers {
		headerData[i] = h
	}
	if err := sw.SetRow(headerCell, headerData); err != nil {
		log.Fatalf("写入表头失败: %v", err)
	}
	r.nextRow = 2
	for i := 0; i < len(data); i++ {
		cell, _ := excelize.CoordinatesToCellName(1, r.nextRow)
		err := sw.SetRow(cell, data[i])
		if err != nil {
			fmt.Printf("追加数据失败（行%v列%v）：%v", cell, data[i], err)
			panic(err)
		}
		r.nextRow++
	}
	return f, sw
}
func (r *ExportService) appendDataToExcel(data [][]interface{}, sw *excelize.StreamWriter) {
	for i := 0; i < len(data); i++ {
		cell, _ := excelize.CoordinatesToCellName(1, r.nextRow)
		err := sw.SetRow(cell, data[i])
		if err != nil {
			fmt.Printf("追加数据失败（行%v列%v）：%v", cell, data[i], err)
			panic(err)
		}
		r.nextRow++
	}
}

func (r *ExportService) getDataTotalNum(baseSql string) error {
	countSql := r.ParseSQLToHeaderAndCountSql(baseSql)
	err := r.Db.Raw(countSql).Scan(&r.TotalNum)
	if err != nil {
		return err.Error
	}
	return nil
}

func (r *ExportService) ParseSQLToHeaderAndCountSql(sql string) string {
	selectIdx := strings.Index(sql, "SELECT")
	fromIdx := strings.Index(sql, "FROM")
	if selectIdx == -1 || fromIdx == -1 || selectIdx > fromIdx {
		return ""
	}
	fieldsSubStr := sql[selectIdx+6 : fromIdx]
	fieldsSubStr = strings.TrimSpace(fieldsSubStr)

	fieldList := strings.Split(fieldsSubStr, ",")
	header := make([]string, 0, len(fieldList))
	for _, field := range fieldList {
		cleanField := strings.TrimSpace(field)
		if cleanField != "" {
			header = append(header, cleanField)
		}
	}
	fromContent := sql[fromIdx:]
	countSql := "SELECT COUNT(*) AS total_num " + fromContent
	return countSql
}
