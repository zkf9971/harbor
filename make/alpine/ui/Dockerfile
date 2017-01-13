FROM alpine:3.5

RUN mkdir /harbor/
RUN apk add --no-cache --update sed apr-util-ldap unzip curl bash

# Add Consul template
# Releases at https://releases.hashicorp.com/consul-template/

ENV CONSUL_TEMPLATE_VERSION 0.18.0-rc1
ENV CONSUL_TEMPLATE_SHA1 a46d00d71c907c3c40af2b23b89c378debc806084d3b4b60f4161c0988edefb1

RUN curl --retry 7 -Lso /tmp/consul-template.zip "https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip" \
    && echo "${CONSUL_TEMPLATE_SHA1}  /tmp/consul-template.zip" | sha256sum -c \
    && unzip /tmp/consul-template.zip -d /usr/local/bin \
    && rm /tmp/consul-template.zip



COPY ./make/dev/ui/harbor_ui /harbor/

COPY ./src/ui/views /harbor/views
COPY ./src/ui/static /harbor/static
COPY ./src/favicon.ico /harbor/favicon.ico
COPY ./make/jsminify.sh /tmp/jsminify.sh

COPY ./make/common/templates/ui/app.conf.template /etc/ui/app.conf.template
COPY ./make/common/templates/ui/private_key.pem /etc/ui/private_key.pem
COPY ./make/common/templates/ui/root.crt /etc/ui/root.crt
COPY ./make/common/templates/ui/start-with-consul.sh /usr/local/bin

RUN chmod u+x /harbor/harbor_ui \
    && timestamp=`date '+%s'` \
    && /tmp/jsminify.sh /harbor/views/sections/script-include.htm /harbor/static/resources/js/harbor.app.min.$timestamp.js /harbor/ \
    && sed -i "s/harbor\.app\.min\.js/harbor\.app\.min\.$timestamp\.js/g" /harbor/views/sections/script-min-include.htm \
    && echo "TLS_REQCERT allow" >> /etc/openldap/ldap.conf
	
EXPOSE 9000

WORKDIR /harbor/
#ENTRYPOINT ["/harbor/harbor_ui"]
