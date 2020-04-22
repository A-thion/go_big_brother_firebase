FROM ubuntu

ENV PORT="8080"

COPY firebase_key.json .
COPY main_firebase .

RUN apt-get update \
 && apt-get install -y --force-yes --no-install-recommends \
      apt-transport-https \
      curl \
      ca-certificates \
 && apt-get clean \
 && apt-get autoremove \
 && rm -rf /var/lib/apt/lists/*
RUN chmod +x ./main_firebase

CMD ["/main_firebase"]