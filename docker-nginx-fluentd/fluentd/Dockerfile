FROM fluent/fluentd:v0.12-onbuild
LABEL maintainer="Yuankun Zhang <i@yuankun.me>"

RUN apk add --update --virtual .build-deps sudo build-base ruby-dev \
    && sudo gem install fluent-plugin-s3 \
    && sudo gem install fluent-plugin-elasticsearch \
    && sudo gem sources --clear-all \
    && apk del .build-deps \
    && rm -rf /var/cache/apk/* /home/fluent/.gem/ruby/2.3.0/cache/*.gem
