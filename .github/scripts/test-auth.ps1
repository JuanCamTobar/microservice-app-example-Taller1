Write-Host "Testing Auth API /version..."
try {
  $response = Invoke-RestMethod http://localhost:8080/version
  if ($response.Trim() -ne "Auth API, written in Go") {
    Write-Error "Auth API test failed: got '$response'"
    exit 1
  }
  Write-Host "Auth API test passed"
} catch {
  Write-Error "Auth API test failed: $_"
  exit 1
}
