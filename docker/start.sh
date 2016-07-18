#!/bin/sh

echo configuring htpasswd
if [ $USE_BASICAUTH == "true" ]; then
  htpasswd -b -c /etc/nginx/.htpasswd $BASICAUTH_USERNAME $BASICAUTH_PASSWORD
  cp ./docker/nginx_default_basicauth /etc/nginx/sites-enabled/default
else
  cp ./docker/nginx_default /etc/nginx/sites-enabled/default
fi

echo configuring nginx
sed -i "s/{SERVER_NAME}/$SERVER_NAME/g" /etc/nginx/sites-enabled/default

echo starting khan
/usr/bin/supervisord --nodaemon --configuration /etc/supervisord-khan.conf
