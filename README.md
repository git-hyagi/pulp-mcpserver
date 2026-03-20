# pulp-mcpserver


## configuring
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

## prompt samples
  * onboard me to pulp rpm in test domain
  * create a pulp rpm repository called test
  * list all python distributions in test domain