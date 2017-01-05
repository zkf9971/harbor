echo "checking $CONSUL:8500..."

wait_for_port() {
    local HOST=$1
    local PORT=$2
    while ! timeout --preserve-status 2 bash -c "</dev/tcp/${HOST}/${PORT}"; do
        echo "wait for 3 more seconds for ${HOST}:${PORT} to be ready...";
        sleep 3;
    done
}

wait_for_port $CONSUL 8500

consul-template \
    -log-level debug \
    -once \
    -dedup \
    -consul ${CONSUL}:8500 \
    -template "/etc/ui/app.conf.template:/etc/ui/app.conf"
    $@