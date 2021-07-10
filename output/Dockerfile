FROM volcanosh/vc-scheduler:v1.2.0

COPY plugins plugins

RUN adduser -h /home/hwMindX -u 9000 -D -s /sbin/nologin hwMindX && \
    chown -R hwMindX:hwMindX plugins /vc-scheduler && \
    chmod -R 550 plugins /vc-scheduler

USER hwMindX