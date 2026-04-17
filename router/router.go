package router

import (
	"main/agent"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type IRouter interface {
	Register(engine *gin.Engine)
	SpanFilter(r *gin.Context) bool
	AccessRecordFilter(r *gin.Context) bool
}

type Router struct {
	rootPath string
	agentApi *agent.Agent
}

func NewRouter(rootPath string, agentApi *agent.Agent) *Router {
	return &Router{rootPath: rootPath,
		agentApi: agentApi}
}

func (r *Router) Register(engine *gin.Engine) {
	root := engine.Group(r.rootPath)
	r.route(root)
}
func (r *Router) SpanFilter(c *gin.Context) bool {
	return false
}
func (r *Router) AccessRecordFilter(c *gin.Context) bool {
	return false
}

func (r *Router) route(root *gin.RouterGroup) {

	api := root.Group("/api")
	{
		api.GET("/chat/history", r.agentApi.GetHis)
		api.GET("/stream", r.agentApi.StreamHandler)
	}
	staticHomeDir, _ := filepath.Abs("./static/home")
	root.Static("/static", staticHomeDir)
}
