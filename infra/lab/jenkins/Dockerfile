# Лабораторный Jenkins: docker CLI + curl; root — чтобы не подбирать GID docker.sock на хосте.
FROM jenkins/jenkins:lts-jdk17
USER root
RUN apt-get update \
  && apt-get install -y --no-install-recommends docker.io ca-certificates curl \
  && rm -rf /var/lib/apt/lists/*
# Только для lab; в проде используйте group_add и пользователя jenkins.
USER root
