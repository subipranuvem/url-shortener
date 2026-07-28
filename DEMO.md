# Demo: Cluster ScyllaDB Distribuído

Cada máquina vira um nó real do cluster. O professor sobe o seed + API completa (3 nós locais). Cada aluno sobe um nó ScyllaDB que entra no cluster.

> **Pré-requisito de rede**: professor e alunos devem estar na **mesma rede WiFi**. O cluster usa o IP local (LAN) — IPs de redes distintas não se comunicam via gossip.

---

## Pré-requisitos (todos)

- Docker + Docker Compose instalados
- Clonar este repositório

---

## 1. Professor: descobrir o IP local

```bash
ipconfig getifaddr en0
# Exemplo: 192.168.1.100
```

> Esse é o `TEACHER_IP`. Use sempre o IP da interface WiFi (`en0`), não o IP NetBird nem o Docker interno.

---

## 2. Professor: subir o cluster seed + API

```bash
export TEACHER_IP=$(ipconfig getifaddr en0)

docker compose -f docker-compose.teacher.yml up -d
```

Acompanhar a subida:
```bash
docker compose -f docker-compose.teacher.yml logs -f scylla-seed scylla-2 scylla-3
```

Aguardar os 3 nós ficarem `healthy` (~3-4 min) antes de liberar para os alunos:
```bash
docker compose -f docker-compose.teacher.yml ps
```

---

## 3. Aluno: entrar no cluster

```bash
# No diretório do projeto clonado
export TEACHER_IP=<IP_LAN_DO_PROFESSOR>   # ex: 192.168.1.100
export MY_IP=$(ipconfig getifaddr en0)    # IP LAN do aluno (macOS)
# Linux: export MY_IP=$(hostname -I | awk '{print $1}')

docker compose -f docker-compose.student.yml up -d
```

O nó entra em bootstrap e começa a receber schema do seed. Demora ~2 minutos.

---

## 4. Verificar o cluster

No terminal do professor:
```bash
docker exec scylla-seed nodetool status
```

Saída esperada (professor + 2 alunos):
```
Datacenter: datacenter1
=======================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address         Load    Tokens Owns  Host ID  Rack
UN 192.168.1.100  123 KB  256    ?     ...      rack1   <- professor (seed)
UN 192.168.1.100  118 KB  256    ?     ...      rack1   <- professor (scylla-2)
UN 192.168.1.100  115 KB  256    ?     ...      rack1   <- professor (scylla-3)
UN 192.168.1.101  98 KB   256    ?     ...      rack1   <- aluno 1
UN 192.168.1.102  87 KB   256    ?     ...      rack1   <- aluno 2
```

`UN` = Up + Normal. `UJ` = ainda em bootstrap, aguardar.

---

## 5. Acessar a API

A API roda na máquina do professor. Alunos na mesma rede acessam pelo IP do professor:

- KrakenD (gateway): `http://<TEACHER_IP>:8000`
- API direta: `http://<TEACHER_IP>:8080`

```bash
curl -X POST http://<TEACHER_IP>:8000/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

---

## Comandos úteis

```bash
# Ver status do cluster (professor)
docker exec scylla-seed nodetool status

# Ver distribuição de tokens
docker exec scylla-seed nodetool ring

# Ver logs do nó (aluno)
docker logs scylla-node --follow

# Simular falha de nó (aluno)
docker compose -f docker-compose.student.yml stop scylla-node

# Recolocar nó no ar
docker compose -f docker-compose.student.yml start scylla-node

# Limpar tudo (aluno)
docker compose -f docker-compose.student.yml down -v

# Limpar tudo (professor)
docker compose -f docker-compose.teacher.yml down -v
```

---

## Troubleshooting

### Nó do aluno fica em `UJ` (joining) por mais de 5 min

```bash
# Testar conectividade com o seed do professor
nc -zv <TEACHER_IP> 7000

# Se falhar: máquinas em redes diferentes ou firewall bloqueando porta 7000
# Confirmar que ambos estão na mesma WiFi
```

### Erro `broadcast address mismatch`

`MY_IP` não é o IP da interface de rede correta. Verificar:
```bash
ifconfig | grep "inet " | grep -v 127
# Usar o IP da interface WiFi (en0 no macOS)
```

### Nós 2 e 3 do professor não ficam healthy

Verificar se `TEACHER_IP` está correto:
```bash
echo $TEACHER_IP
# Deve ser o IP WiFi (192.168.x.x), não IP Docker (172.x.x.x) nem NetBird (100.x.x.x)
```

### Container `sysctl-init` falha

Em macOS com Docker Desktop: ignorar, o Docker Desktop gerencia o kernel automaticamente.
Em Linux: rodar com `sudo docker compose ...` ou adicionar o usuário ao grupo `docker`.

### API não conecta ao ScyllaDB

A API conecta ao `scylla-seed` via rede interna Docker. Verificar:
```bash
docker logs url-shortener
docker logs migrate
```

### Fator de replicação insuficiente

Escritas com `CONSISTENCY=QUORUM` exigem `RF >= 3` e pelo menos 2 nós disponíveis. Com menos de 3 nós no total:

```bash
docker exec -it scylla-seed cqlsh
ALTER KEYSPACE shortener WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 2};
```

---

## Acesso remoto (opcional)

Para alunos em redes diferentes acessarem **somente a API** (sem entrar no cluster como nó):

1. Professor instala e sobe NetBird: `sudo netbird up`
2. Professor expõe a porta da API: `ngrok http 8000`
3. Alunos usam a URL do ngrok — não precisam de Docker nem do `docker-compose.student.yml`

Neste caso os alunos são **clientes** da API, não nós do cluster.
