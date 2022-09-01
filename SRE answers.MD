# SRE Answers

1. Nginx раздаёт видео файлами mp4 по byte-range. Нужно написать
   конфиг, который ограничивает скорость потока в 5 Мбит. Видео
   доступно по временной ссылке в течение одного часа. По
   истечении часа ссылка должна возвращать код 410.

   честно сказать никогда не работал с тем чтобы использовать эти фичи nginx, 
   поэтому просто почитал доку и наваял, наверняка можно лучше
```chell
cat < EOF > test.conf
limit_conn_zone $binary_remote_addr zone=limitconnbyaddr:10m;
limit_conn_status 410;

server {
    listen SOME_PORT;
    server_name SOME_SERVER_NAME;

    location  /s/ {
        limit_rate 5m;
        limit_conn   limitconnbyaddr  1;
        limit_conn_status 429;
        secure_link $arg_md5,$arg_expires;
        secure_link_md5 "$secure_link_expires$uri$remote_addr enigma";

        if ($secure_link = "") { return 403; }
        if ($secure_link = "0") { return 410; }
    }
}
EOF
```
2. На сервере с Nginx высокая нагрузка из-за неверного конфига. Мы
   изменили конфиг и пришли к выводу, что нужно раскатать новый
   конфиг и перезапустить все процессы Nginx с новым конфигом.
   Серверов 20 штук. Напишите команду.

   1. If it's a systemd service: `clush <(cat /list/of/all/servers) sudo systemctl reload nginx` or pssh or even ansible. whatever suits
   2. if It's something else, but you know nginx path: `clush <(cat /list/of/all/servers) nginx -s reload`
   3. if you don't know nginx path, or it's different from server to server - you are awful at your job ¯\_(ツ)_/¯
3. Идёт большой поток запросов на Nginx приводящий к ошибкам
   410/403/404. Нужно кикнуть айпишники, которые генерируют
   больше 10 запросов с ошибками за минуту в бан на 15 минут.
   Решите любым удобным инструментом.

   fail2ban 
```shell
sed -i -e 's/bantime = */bantime = 600/' -e 's/findtime = */findtime = 60/' -e 's/maxretry = */maxretry = 10/' /etc/fail2ban/jail.local # но обычно там и так 10 минут дефолт бана
cat << EOF > /etc/fail2ban/filter.d/nginx-400.conf
[Definition]

# наверное надо бы получше написать, но откуда мне знать какой формат логов настроен
failregex = ^<HOST> -.*GET.*" (410|403|404) .*
EOF

cat << EOF >> /etc/fail2ban/jail.local
[nginx-400]

enabled  = true
filter   = nginx-400
port     = http,https
logpath  = /var/log/nginx/access.log
EOF
```
4. На некоторых серверах отвалился mount point. Нужно
   перемонтировать. Перемонтирование может не сработать с
   первого раза. Серверов более 100.
```shell
ansible-galaxy collection install ansible.posix
cat << EOF > mount.yaml
- name: Mount DVD read-only
  ansible.posix.mount:
    path: /some/path
    src: /some/src
    fstype: whatever
    opts: ro,noauto
    state: present
  retries: 3
  delay: 3
  register: result
  until: result is not failed
EOF
ansible-playbook -i /your/inventory/with/more/than/100/servers.ini mount.yaml
```
5. MySQL работает в master-slave конфигурации. Каждый день
   репликация останавливается. В статусе видим:
   Last_SQL_Errno: 1205
   Last_SQL_Error: Lock wait timeout exceeded; try restarting transaction
   Какие могут быть причины?

   скорее всего у вас проблема в том, что выбрали мусю вместо постгри. 
   Поставили разные таймауты на innodb_lock_wait_timeout на мастере и слейве и сейчас ловите эти ошибки.
   С постгресом было бы проще

6. Нужно скопировать директорию с большим количеством мелких
   файлов с сервера A на сервер B. Ключи ssh к обоим находятся
   только у вас на компьютере. Вход только по ключам. Написать
   команду в одну строку.
```shell
tar --use-compress-program="pigz -k " cf - /some/path/to/folder | ssh user@host "cd /path/on/the/remote && tar --use-compress-program='pigz -d' xvf -"
```
7. На сервере 4 винта: sda, sdb, sdc, sdd. Каждый разбит на 4 раздела.
   1-й в md0 raid1 /boot
   2-й в md1 raid0 swap
   3-й в md2 raid6 /
   4-й в xfs и смонтирован в /mnt/sd?4
   Винт sdd вышел из строя, его заменили на новый. Включите его в
   систему (набор команд) без перезагрузки сервера.
```shell
#create partitions with fdisk or parted
mdadm --manage /dev/md0 -a /dev/sdd1
mdadm --manage /dev/md1 -a /dev/sdd2
mdadm --manage /dev/md2 -a /dev/sdd3
mkfs.xfs /dev/sdd4
mount /dev/sdd4 /mnt/sdd4
```
