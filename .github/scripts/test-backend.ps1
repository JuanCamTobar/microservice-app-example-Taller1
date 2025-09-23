foreach ($port in 8082,8083) {
  Write-Host "Testing health endpoint on port $port..."
  try {
    $response = Invoke-RestMethod http://localhost:$port/health
    if ($response.status -ne "ok") {
      Write-Error "Health check failed on port ${port}: got '$($response | ConvertTo-Json)'"
      exit 1
    }
    Write-Host "Health check passed on port $port"
  } catch {
    Write-Error "Health check failed on port ${port}: $_"
    exit 1
  }
}
