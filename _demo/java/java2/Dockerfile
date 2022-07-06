FROM adoptopenjdk/openjdk11:alpine-jre
ARG JAR_FILE=target/*.jar
COPY ${JAR_FILE} application2.jar
EXPOSE 8081
ENTRYPOINT ["java","-jar","/application2.jar"]