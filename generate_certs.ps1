$ErrorActionPreference = "Stop"

Write-Host "Generating PKI certificates"

Set-Location $PSScriptRoot

# Create Root CA
openssl req -x509 -newkey rsa:2048 -nodes `
    -keyout certs/ca.key `
    -out certs/ca.crt `
    -days 365 `
    -subj "/CN=MyRootCA"

# Create server private key and CSR
openssl req -newkey rsa:2048 -nodes `
    -keyout certs/server.key `
    -out certs/server.csr `
    -subj "/CN=localhost"

# Create SAN configuration
@"
subjectAltName=DNS:localhost,IP:127.0.0.1
"@ | Out-File -Encoding ascii certs/server_ext.cnf

# Sign server certificate
openssl x509 -req `
    -in certs/server.csr `
    -CA certs/ca.crt `
    -CAkey certs/ca.key `
    -CAcreateserial `
    -out certs/server.crt `
    -days 365 `
    -extfile certs/server_ext.cnf

# Create client private key and CSR
openssl req -newkey rsa:2048 -nodes `
    -keyout certs/client.key `
    -out certs/client.csr `
    -subj "/CN=dispatcher-agent-01"

# Sign client certificate
openssl x509 -req `
    -in certs/client.csr `
    -CA certs/ca.crt `
    -CAkey certs/ca.key `
    -CAcreateserial `
    -out certs/client.crt `
    -days 365

Write-Host ""
Write-Host "=== Certificates generated successfully ==="

Get-ChildItem certs