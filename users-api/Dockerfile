# ============================
# Stage 1: Build con Maven
# ============================
FROM maven:3.9.6-eclipse-temurin-8 AS builder

# Evita descargar dependencias en cada build
WORKDIR /build
COPY pom.xml ./
RUN mvn -B -ntp dependency:go-offline

# Copia código y empaqueta (fat JAR de Spring Boot)
COPY src ./src
RUN mvn -B -ntp -DskipTests package

# ============================
# Stage 2: Runtime (JRE 8)
# ============================
FROM eclipse-temurin:8-jre

# Opcional: flags JVM (memoria, GC, etc.)
ENV JAVA_OPTS=""

WORKDIR /app
# Copia el jar generado (target/<app>.jar) y lo renombra
COPY --from=builder /build/target/*.jar /app/app.jar

# El Users API escucha en 8083 en este proyecto
EXPOSE 8083

# Ejecuta como usuario sin privilegios (uid arbitrario)
USER 10001

# Arranque
ENTRYPOINT [ "sh", "-c", "java $JAVA_OPTS -jar /app/app.jar" ]