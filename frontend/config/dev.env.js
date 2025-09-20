var merge = require('webpack-merge')
var prodEnv = require('./prod.env')

module.exports = merge(prodEnv, {
  NODE_ENV: '"development"',
  AUTH_API_ADDRESS: '"http://localhost:8080"',
  TODOS_API_ADDRESS: '"http://localhost:8082"',
  ZIPKIN_URL: '"http://localhost:9411/api/v2/spans"',
})
