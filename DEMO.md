# Demo: Cluster ScyllaDB Distribuído

O professor sobe o nó seed + API. Dois alunos entram como nós 2 e 3. Com 3 nós no cluster, o professor roda a migration e a API fica operacional.

> **Pré-requisito de rede**: professor e alunos devem estar na **mesma rede WiFi**. O cluster usa IP local (LAN).

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

> Esse é o `TEACHER_IP`. Use o IP da interface WiFi (`en0`). Não use IP Docker (`172.x`) nem NetBird (`100.x`).

---

## 2. Professor: subir seed + API stack

```bash
export TEACHER_IP=$(ipconfig getifaddr en0)

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
export TEACHER_IP=<IP_LAN_DO_PROFESSOR>   # ex: 192.168.1.100
export MY_IP=$(ipconfig getifaddr en0)    # macOS
# Linux: export MY_IP=$(hostname -I | awk '{print $1}')

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
-- Address         Load    Tokens Owns  Host ID  Rack
UN 192.168.1.100  123 KB  256    ?     ...      rack1   <- professor (seed)
UN 192.168.1.101  98 KB   256    ?     ...      rack1   <- aluno 1
UN 192.168.1.102  87 KB   256    ?     ...      rack1   <- aluno 2
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
# Testar conectividade com o seed do professor
nc -zv <TEACHER_IP> 7000

# Se falhar: máquinas em redes diferentes ou firewall bloqueando porta 7000
# Confirmar que ambos estão na mesma WiFi
```

### Erro `broadcast address mismatch`

`MY_IP` não é o IP da interface correta:
```bash
ifconfig | grep "inet " | grep -v 127
# Usar o IP da interface WiFi (en0 no macOS)
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

## Acesso remoto (opcional)

Para alunos em redes diferentes acessarem **somente a API** (sem entrar no cluster como nó):

1. Professor instala e sobe NetBird: `sudo netbird up`
2. Professor expõe a porta da API: `ngrok http 8000`
3. Alunos usam a URL do ngrok — não precisam de Docker

Neste caso os alunos são **clientes** da API, não nós do cluster.
