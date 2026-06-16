@echo off
cd /d D:\k8s-go\k8s_operation
echo Starting backend at %date% %time% > backend_start.log
go run ./cmd/k8soperation >> backend_start.log 2>&1
