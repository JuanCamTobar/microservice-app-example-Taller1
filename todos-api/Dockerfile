# Stage 1: Build
FROM node:8.17.0 AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install --legacy-peer-deps
COPY . .

# Stage 2: Runtime
FROM node:8.17.0-alpine
WORKDIR /app
COPY --from=builder /app .
EXPOSE 8082
ENV TODO_API_PORT=8082
CMD ["npm", "start"]