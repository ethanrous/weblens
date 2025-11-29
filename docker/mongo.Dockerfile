ARG ARCHITECTURE

FROM --platform=linux/${ARCHITECTURE} mongo

ENV WEBLENS_MONGO_HOST_NAME=weblens-mongo

HEALTHCHECK --retries=16 --interval=5m --start-period=5s --start-interval=2s --timeout=4s CMD mongosh --eval "try { rs.status() } catch (err) { rs.initiate({_id:'rs0',members:[{_id:0,host:'$WEBLENS_MONGO_HOST_NAME:27017'}]}) }" 


ENTRYPOINT ["mongod"]
CMD ["--replSet", "rs0", "--bind_ip_all"]
