# Lessons Learned

Problems and resolutions encountered during development of this project.

---

## ScyllaDB

### `deploy.resources.limits` not applied as cgroup limits

`deploy.resources.limits` under `docker compose up` is ignored — it only takes effect with Docker Swarm (`docker stack deploy`). ScyllaDB's Seastar engine reads cgroup memory limits to auto-configure itself, so it was seeing the full host memory and failing with "insufficient physical memory" when the VM had less than expected.

**Fix:** use `mem_limit:` at the service level, which is the Compose v2 equivalent that Docker actually enforces.

---

### AIO slots exhausted with `--smp 2`

Seastar (ScyllaDB's engine) requires async I/O slots from the Linux kernel. Each shard needs ~66k slots. Docker Desktop's Linux VM defaults to 65536 (`aio-max-nr`), which is enough for one shard but causes a FATAL error on startup with 2+ shards.

**Fix:** added a `sysctl-init` privileged container that runs before the Scylla nodes and sets `fs.aio-max-nr=1048576`. This resets on Docker Desktop restart, but re-running `docker compose up` reapplies it automatically.

---

### Stale gossip state after container restart

Nodes 2 and 3 failed on restart with "Attempt to add X as saved endpoint" because old node state persisted on the volumes.

**Fix:** `docker compose down -v` to wipe volumes before bringing the cluster back up. Added `make docker-reset` for this.

---

### `USE` statements not supported by gocql

gocql does not allow `USE keyspace` statements — they are rejected at the driver level.

**Fix:** the migration runner detects `USE keyspace` statements, closes the current session, sets `cluster.Keyspace`, and reconnects. Subsequent statements in that migration file run under the correct keyspace.

---

### Write latency under high load (connection reset)

With `Consistency = Quorum` and RF=3, every write requires 2/3 nodes to acknowledge before returning. Under high concurrency, slow nodes cause the coordinator to wait, the API response is delayed, KrakenD times out, and the client gets a connection reset.

**Fix:** changed consistency to `ONE` — the write returns after 1 node acknowledges; the other 2 replicate asynchronously. Also increased KrakenD's backend timeout from 3s to 10s.

---

## KrakenD

### `include` does not resolve template variables

KrakenD's `FC_PARTIALS` with `include` performs verbatim text substitution. Template variables like `{{ .api_config.api_host }}` inside an included file are not evaluated — they render as `<no value>`, producing invalid JSON.

**Fix:** moved all endpoint definitions that reference settings variables to `FC_TEMPLATES` using `{{ define "name" }}` / `{{ template "name" . }}` blocks, which receive the full template context.

---

### Templates directory must exist on disk

If `FC_TEMPLATES` points to a directory that doesn't exist, KrakenD fails at startup — even if no templates are actually used yet.

**Fix:** added `.gitkeep` to `krakend/templates/` so the directory is always present.

---

## Docker / Docker Compose

### Swagger UI self-copy error

Mounting `openapi.yaml` directly into `/usr/share/nginx/html/` caused nginx's startup script to try copying the file over itself.

**Fix:** mount to `/tmp/openapi.yaml` and set `SWAGGER_JSON=/tmp/openapi.yaml`.

---

### `migrate` container hanging

Using the ScyllaDB image to run `cqlsh` scripts via a separate entrypoint didn't work — the image entrypoint starts the Scylla server, not a shell.

**Fix:** wrote a standalone Go binary (`cmd/migrate`) compiled into a minimal Alpine image (`Dockerfile.migrate`). It connects via gocql and executes migration files directly.

---

## Go HTTP Client

### TCP connection not reused — TIME_WAIT exhaustion

Go's `net/http` Transport only returns a connection to the idle pool if the response body is **fully read** before `Close()` is called. In the seed's success path (HTTP 201), the body was closed without being read, so every successful request dropped the TCP connection. With ~1300 req/s, TIME_WAIT sockets accumulated faster than macOS's ~16k ephemeral ports could handle, causing `connect: resource temporarily unavailable`.

**Fix:** always read and discard the full response body (`io.ReadAll(io.LimitReader(...))`) before closing, for all response status codes.

---

## Alias Generation

### Birthday paradox with small combination space

With 100 adjectives × 100 names = 10,000 combinations, collision probability becomes very high well before 10,000 inserts (birthday paradox). At 100k requests, collision rate was unacceptable.

**Fix:** added a third factor (animals), expanding the space to 224 × 194 × 260 ≈ 11.3M combinations.


----


Adicionar isso depois

Alguns comandos para cadastrar na documentação


Pra ver quantos cores o servidor tem:

docker exec -it scylla-node3 bash

./usr/lib/scylla/seastar-cpu-map.sh -n scylla

---

Parametro aio-max-nr

Before starting the cluster, make sure the aio-max-nr value is high enough (1048576 or more). 

This parameter determines the maximum number of allowable Asynchronous non-blocking I/O (AIO) concurrent requests by the Linux Kernel, and it helps ScyllaDB perform in a heavy I/O workload environment.


Check the value: 
cat /proc/sys/fs/aio-max-nr

If it needs to be changed:
echo "fs.aio-max-nr = 1048576" >> /etc/sysctl.conf

---

To see the tokens for each node on our cluster, we use the nodetool ring command:

docker exec -it scylla-node1 nodetool ring
This shows us the token range for each node. The value in the column Token is the end of the token range, up to (and including) the value listed.

Another way to show the tokens present on a specific node is with the describing command using the keyspace as a parameter:

docker exec -it scylla-node1 nodetool describering scyllau

---

To see whether a node is communicating using Gossip, we use the statusgossip command:

docker exec -it scylla-node1 nodetool statusgossip
We can see that scylla-node1 is communicating with other nodes. This communication is on by default unless it has been turned off, for example, for maintenance.

To see what a node is communicating to the other nodes in the cluster, we use the gossipinfo command:

docker exec -it scylla-node1 nodetool gossipinfo

---

To see the snitch defined in our multi-dc cluster, we use the describecluster command:

docker exec -it scylla-node1 nodetool describecluster
We can see that the Snitch being used is the GossipingPropertyFileSnitch. It allows us to explicitly define which DC and Rack a specific Node belongs to. This snitch reads its configuration from a cassandra-rackdc.properties file located under /etc/scylla/

docker exec -it  scylla-node1 cat /etc/scylla/cassandra-rackdc.properties

---

