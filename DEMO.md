# Demo: Cluster ScyllaDB Distribuído

O professor sobe o nó seed + API. Dois alunos entram como nós 2 e 3. Com 3 nós no cluster, o professor roda a migration e a API fica operacional.

> **Pré-requisito de rede**: professor e alunos devem ter o **NetBird instalado e conectado na mesma rede virtual**. O gossip do ScyllaDB usa os IPs NetBird (`100.x.x.x`).

---

## Pré-requisitos (todos)

- Docker + Docker Compose instalados
- Clonar este repositório
- NetBird instalado e conectado: `sudo netbird up`

---

## 1. Professor: descobrir o IP NetBird

```bash
netbird status | grep "NetBird IP"
# Exemplo: NetBird IP: 100.64.0.1/24
```

> Esse é o `TEACHER_IP`. Use o IP NetBird (`100.x.x.x`). Não use IP WiFi LAN (`192.168.x`) nem IP Docker (`172.x`).

---

## 2. Professor: subir seed + API stack

```bash
export TEACHER_IP=$(netbird status | grep "NetBird IP:" | awk '{print $3}' | cut -d'/' -f1)

docker compose -f docker-compose.teacher.yml up -d scylla-seed redis krakend
```

Aguardar o seed ficar `healthy` (~90s):
```bash
docker compose -f docker-compose.teacher.yml ps scylla-seed
```

**Não rodar a API ainda** — aguardar alunos entrarem primeiro.

---

## 3. Aluno: entrar no cluster

```bash
# No diretório do projeto clonado
export TEACHER_IP=<IP_NETBIRD_DO_PROFESSOR>   # ex: 100.64.0.1

# Ubuntu (NetBird cria interface wt0):
export MY_IP=$(netbird status | grep "NetBird IP:" | awk '{print $3}' | cut -d'/' -f1)
# ou: MY_IP=$(ip addr show wt0 | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1)

docker compose -f docker-compose.student.yml up -d
```

O nó entra em bootstrap e recebe schema do seed. Demora ~2 minutos.

Verificar se entrou (`UJ` = ainda bootstrapping, `UN` = pronto):
```bash
docker logs scylla-node --follow
```

---

## 4. Professor: verificar 3 nós e rodar migration

Após os 2 alunos entrarem, verificar:
```bash
docker exec scylla-seed nodetool status
```

Saída esperada (seed + 2 alunos):
```
Datacenter: datacenter1
=======================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address       Load    Tokens Owns  Host ID  Rack
UN 100.64.0.1   123 KB  256    ?     ...      rack1   <- professor (seed)
UN 100.64.0.2   98 KB   256    ?     ...      rack1   <- aluno 1
UN 100.64.0.3   87 KB   256    ?     ...      rack1   <- aluno 2
```

Com 3 nós `UN`, rodar a migration:
```bash
docker compose -f docker-compose.teacher.yml run --rm migrate
```

---

## 5. Professor: subir a API

```bash
docker compose -f docker-compose.teacher.yml up -d api
```

Verificar:
```bash
docker logs url-shortener --follow
```

---

## 6. Acessar a API

Todos os alunos acessam pelo IP do professor:

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

# Simular falha de nó (aluno para o container)
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

### Nó do aluno fica em `UJ` por mais de 5 min

```bash
# Testar conectividade com o seed do professor via NetBird
nc -zv <TEACHER_IP> 7000

# Se falhar: NetBird desconectado ou firewall bloqueando porta 7000
netbird status   # confirmar "Connected"
```

### Erro `broadcast address mismatch`

`MY_IP` não é o IP NetBird:
```bash
netbird status | grep "NetBird IP"
# Confirmar que MY_IP bate com esse valor
```

### Migration falha com "not enough nodes"

Menos de 3 nós no cluster quando a migration rodou. Aguardar os 2 alunos entrarem (`UN`) antes de rodar:
```bash
docker exec scylla-seed nodetool status
# Confirmar 3 linhas UN antes de rodar migrate
```

### Container `sysctl-init` falha

macOS Docker Desktop: ignorar, gerenciado automaticamente.
Linux: rodar com `sudo` ou adicionar usuário ao grupo `docker`.

### API não conecta ao ScyllaDB

Migration ainda não rodou ou falhou:
```bash
docker logs url-shortener
docker compose -f docker-compose.teacher.yml run --rm migrate
```

---

## Acesso remoto (somente API, sem entrar no cluster)

Para alunos que querem apenas consumir a API sem virar nó:

1. Professor e aluno conectados no NetBird
2. Aluno acessa direto pelo IP NetBird do professor: `http://<TEACHER_NETBIRD_IP>:8000`

Sem Docker necessário no lado do aluno.

### Alternativa sem NetBird (ngrok)

1. Professor: `ngrok http 8000`
2. Alunos usam a URL do ngrok

Neste caso os alunos são **clientes** da API, não nós do cluster.
