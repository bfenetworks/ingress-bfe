FROM bfenetworks/bfe



COPY dist/start.sh /bfe/bin/
COPY dist/bfe_ingress_controller /bfe/bin/
RUN chmod u+x /bfe/bin/start.sh


WORKDIR /bfe/bin/
EXPOSE 8080 8443 8421

ENTRYPOINT ["./start.sh"]