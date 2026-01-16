# Eino-example

这是一个基于 字节跳动开源的大语言模型应用框架 **Eino** 构建 **RAG检索增强生成** 的简单示例。</br>
本示例比官方示例更加通俗，摒弃了官方示例的冗杂向量数据库配置内容，选择**redis**作为向量数据库，更适合新手及小型项目。

## 什么是 RAG ？
**RAG**（Retrieval-Augmented Generation，**检索增强生成**）是一种结合**信息检索技术**和**语言生成模型**的人工智能技术。
它通过从外部知识库中检索相关信息，并将其作为提示输入给大型语言模型（LLMs），从而增强模型处理知识密集型任务的能力。

## 内容介绍：
- **RAGEngine**： RAG引擎，用于向量的嵌入、本地存储、检索、文档生成。
- **docker-compose.yaml**： 用于构建 **redis-stack** 的 **docker** 环境，本项目的向量数据库依赖于 **redis-stack**。
- **test-txt**: 用于向量化的文档。
- **config**： 一些隐私性配置内容，如**APIKey**。
- **其他内容**： RAG引擎的各个组件，如 loader、splitter、retriever等，均为 Eino框架 的重要构成**组件 (Components)**。


**具体内容请参考代码内容！**</br>
**don't speak and show you the codes!**