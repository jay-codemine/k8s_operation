FROM busybox

RUN mkdir -p /app/configs
COPY bin/devops-be /app/devops-be
COPY configs/* /app/configs/
RUN chmod +x /app/devops-be

EXPOSE 8080

CMD ["/app/devops-be"]
