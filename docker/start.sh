#!/bin/sh


echo configuring htpasswd
htpasswd -b -c /etc/nginx/.htpasswd $AUTH_USERNAME $AUTH_PASSWORD

echo configuring nginx
sed -i "s/{SERVER_NAME}/$SERVER_NAME/g" /etc/nginx/sites-enabled/default

echo starting khan
/usr/bin/supervisord --nodaemon --configuration /etc/supervisord-khan.conf
