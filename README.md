# pulp-mcpserver


* configuring
```json
    "pulp-mcp": {
      "command": "bash",
      "args": [
        "-c",
        "cd /home/hyagi/pulp/pulp-mcpserver && go run ."
      ],
      "env": {
        "PULP_URL": "http://localhost",
        "PULP_USERNAME": "admin",
        "PULP_PASSWORD": "password"
      }
    }
```
