##################################### stage  1 #####################################
FROM bfe

# adapt BFE conf in docker
COPY output/adapt_bfe_docker.sh /adapt_bfe_docker.sh
RUN sh /adapt_bfe_docker.sh

# pack /home/work directory for COPY in stage 2
COPY build/output/bfe_ingress_controller \
    output/ingress.commit \
    /home/work/go-bfe/bin/
COPY build/output/start.sh /home/work/start.sh

##################################### stage  2 #####################################
# build image from CentOS to reduce image size
FROM centos
RUN yum install -y wget unzip vi net-tools redhat-lsb-core nmap-ncat.x86_64 tcpdump tree jre epel-release \
    && yum install -y supervisor \
    && yum clean all

# copy base files from bfe image
RUN groupadd -g 501 work && useradd -g work -G work -u 500 -d /home/work work
COPY --from=bfe --chown=work:work /home/work /home/work

ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/lib/jvm/jre-openjdk/lib/amd64/server/

RUN mkdir -p /opt/compiler/gcc-4.8.2/lib64 && \
    ln -s /lib64/ld-linux-x86-64.so.2 /opt/compiler/gcc-4.8.2/lib64/ld-linux-x86-64.so.2 && \
    mkdir -p /opt/compiler/gcc-8.2/lib64 && \
    ln -s /lib64/ld-linux-x86-64.so.2 /opt/compiler/gcc-8.2/lib64/ld-linux-x86-64.so.2 && \
    ln -sf /home/work/opbin/pbtool/bin/pblogTool3 /usr/bin/pblogTool3 && \
    ln -s /home/work/opbin/bns_tool_stub /home/work/bns_tool_stub && \
    ln -sf /home/work/opbin/bns_tool_stub/bns_tool_stub /usr/bin/get_instance_by_service && \
    mv /home/work/start.sh /start.sh

WORKDIR /home/work/go-bfe/
USER work
EXPOSE 8080 8443 8421
ENTRYPOINT ["/start.sh"]


