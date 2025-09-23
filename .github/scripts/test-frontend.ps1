Write-Host "Testing frontend on port 8081..."
try {
  $response = Invoke-RestMethod http://localhost:8081 -Method Head -ErrorAction Stop
  Write-Host "Frontend test passed"
} catch {
  Write-Error "Frontend test failed: $_"
  exit 1
}
