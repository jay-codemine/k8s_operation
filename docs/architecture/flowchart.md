# 用户更新流程图

```mermaid
flowchart LR
  A[前端 JSON 请求\nPUT /api/v1/users/:id] --> B[Gin Controller\nShouldBindJSON → DTO]
  B --> C{参数校验\n(必填/格式/业务规则)}
  C -- 通过 --> D[Service 层\n组装更新字段/权限校验]
  C -- 不通过 --> Z[返回 400\n错误信息]
  D --> E[DAO 层\n调用 Model.Update(db, DTO)]
  E --> F[GORM Model(u)\nWhere(id, is_del=0)]
  F --> G{选择 Updates 参数类型}
  G -- DTO(struct) --> H[.Updates(struct)\n默认忽略零值]
  G -- map[string]interface{} --> I[.Updates(map)\n零值也会更新]
  H --> J[(Database)]
  I --> J[(Database)]
  J --> K[返回受影响行数/错误 ]
  K --> L[Controller 返回 200\n标准响应体]
```
