FROM postgres:latest
COPY model.json /docker-entrypoint-initdb.d/
RUN chmod 755 /docker-entrypoint-initdb.d/model.json