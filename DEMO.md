# Demo: Cluster ScyllaDB Distribuído

Cada máquina vira um nó real do cluster. O professor sobe o nó seed + API completa. Cada aluno sobe um único nó ScyllaDB que entra no cluster do professor.

## Pré-requisitos (todos)

- Docker + Docker Compose instalados
- NetBird instalado: https://netbird.io/download
- Conta NetBird (gratuita, até 5 peers na versão cloud)
- Clonar este repositório

## 1. Configurar a rede (NetBird)

### Professor

```bash
# Instalar NetBird (macOS)
brew install netbirdio/tap/netbird

# Subir e autenticar (abre navegador)
sudo netbird up

# Anotar seu IP NetBird
netbird status | grep "NetBird IP"
```

Após autenticar, acesse o dashboard NetBird e gere um **Setup Key** para compartilhar com os alunos.

### Aluno

```bash
# Instalar NetBird (macOS)
brew install netbirdio/tap/netbird

# Linux
curl -fsSL https://pkgs.netbird.io/install.sh | sh

# Entrar na rede do professor com a setup key
sudo netbird up --setup-key <SETUP_KEY_DO_PROFESSOR>

# Anotar seu IP NetBird
netbird status | grep "NetBird IP"
```

## 2. Professor: subir o cluster seed + API

```bash
# No diretório do projeto
export TEACHER_IP=<SEU_IP_NETBIRD>

docker compose -f docker-compose.teacher.yml up -d
```

Aguardar o healthcheck do `scylla-seed` passar (~90s) antes de liberar para os alunos.

Verificar:
```bash
docker logs scylla-seed --follow
```

## 3. Aluno: entrar no cluster

```bash
# No diretório do projeto clonado
export TEACHER_IP=<IP_NETBIRD_DO_PROFESSOR>
export MY_IP=<SEU_IP_NETBIRD>

docker compose -f docker-compose.student.yml up -d
```

O nó vai contactar o seed do professor e entrar no cluster. Pode demorar ~2 minutos.

## 4. Verificar o cluster

No terminal do professor:
```bash
docker exec scylla-seed nodetool status
```

Saída esperada (com 3 alunos + professor):
```
Datacenter: datacenter1
=======================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
-- Address       Load    Tokens Owns    Host ID  Rack
UN 100.64.0.1   123 KB  256    ?       ...      rack1   <- professor
UN 100.64.0.2   98 KB   256    ?       ...      rack1   <- aluno 1
UN 100.64.0.3   87 KB   256    ?       ...      rack1   <- aluno 2
UN 100.64.0.4   91 KB   256    ?       ...      rack1   <- aluno 3
```

`UN` = Up + Normal. Qualquer outro estado: ver seção Troubleshooting.

## 5. Acessar a API

A API roda na máquina do professor. Todos os alunos acessam pelo IP NetBird do professor:

- KrakenD (gateway): `http://<IP_PROFESSOR>:8000`
- API direta: `http://<IP_PROFESSOR>:8080`

Exemplo:
```bash
curl -X POST http://<IP_PROFESSOR>:8000/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

## Comandos úteis

```bash
# Ver logs do nó (aluno)
docker logs scylla-node --follow

# Ver status do cluster (professor)
docker exec scylla-seed nodetool status

# Ver distribuição de tokens
docker exec scylla-seed nodetool ring

# Ver informações de um nó específico
docker exec scylla-seed nodetool info

# Parar nó do aluno (simula falha)
docker compose -f docker-compose.student.yml stop scylla-node

# Subir nó novamente
docker compose -f docker-compose.student.yml start scylla-node

# Limpar tudo (aluno)
docker compose -f docker-compose.student.yml down -v
```

## Troubleshooting

### Nó fica em `joining` por muito tempo

O nó está tentando contatar o seed mas não consegue. Verificar:

```bash
# Testar conectividade com o professor (gossip port)
nc -zv <IP_PROFESSOR> 7000

# Se falhar: NetBird não está ativo ou IP errado
netbird status
```

### Erro `broadcast address mismatch`

`MY_IP` está errado — não é o IP NetBird. Verificar:
```bash
netbird status | grep "NetBird IP"
# Usar exatamente esse IP em MY_IP
```

### Container `sysctl-init` falha

Em macOS com Docker Desktop, ignorar — o Docker Desktop gerencia o kernel.
Em Linux: precisar rodar com `sudo` ou o usuário precisa ter permissão de container privilegiado.

### API não conecta ao ScyllaDB

A API conecta ao `scylla-seed` via rede interna Docker — não via NetBird. Se o container `api` não subiu, verificar:
```bash
docker logs url-shortener
docker logs migrate
```

### Fator de replicação insuficiente

Se o keyspace usa `RF=3` mas há menos de 3 nós no cluster, escritas com `CONSISTENCY=QUORUM` vão falhar. Ajustar:

```bash
# Reduzir RF para o número de nós disponíveis
docker exec -it scylla-seed cqlsh
ALTER KEYSPACE shortener WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 2};
```

Ou subir mais nós de alunos até chegar em 3.
