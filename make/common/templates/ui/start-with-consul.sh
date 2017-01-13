echo "checking $CONSUL:8500..."

while ! echo exit | nc $CONSUL 8500; do echo "wait for 3 more seconds...";sleep 3; done

consul-template \
    -log-level debug \
    -once \
    -dedup \
    -consul ${CONSUL}:8500 \
    -template "/etc/ui/app.conf.template:/etc/ui/app.conf"
    $@