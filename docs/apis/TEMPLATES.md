# API 模板

## 1 OpenAPI 规格模板

```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "Service Name API",
    "version": "1.0.0",
    "description": "API 描述"
  },
  "paths": {
    "/endpoint": {
      "get": {
        "summary": "端点描述",
        "responses": {
          "200": {
            "description": "成功响应"
          }
        }
      }
    }
  }
}
```

## 2 JSON Schema 模板

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "version": "1.0.0",
  "title": "Schema 名称",
  "description": "Schema 描述",
  "type": "object",
  "properties": {
    "field": {
      "type": "string",
      "description": "字段描述"
    }
  },
  "required": ["field"]
}
```
