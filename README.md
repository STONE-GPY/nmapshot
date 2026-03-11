# Nmap Shot

Nmap Shot은 Nmap 스캔을 실행하고 그 결과를 XML 형태로 반환하는 Go 기반의 REST API 서비스입니다. 지정된 타겟과 포트에 대해 스캔을 수행하며, 보안을 위한 API 키 인증 및 허용 포트 설정 기능을 제공합니다.

## 주요 기능

- **Nmap 스캔 실행**: `/api/v1/scan` 엔드포인트를 통해 Nmap 스캔을 수행합니다. (최대 5분 타임아웃 적용)
- **XML 결과 반환**: 스캔 결과를 원시 Nmap XML 형식으로 반환합니다.
- **API 키 인증**: `X-API-Key` 헤더를 통해 모든 스캔 요청을 안전하게 보호합니다.
- **허용 포트 제어**: 환경 변수를 통해 스캔 가능한 포트 또는 포트 범위를 제한할 수 있습니다.
- **Swagger API 문서**: `/docs` 또는 `/swagger` 엔드포인트를 통해 API 명세서를 확인할 수 있습니다.

## 요구 사항

- Go 1.x 이상
- 시스템에 **Nmap** 프로그램이 설치되어 있어야 합니다. 
  - Ubuntu/Debian: `sudo apt install nmap`
  - macOS: `brew install nmap`
  - CentOS/RHEL: `sudo yum install nmap`

## 환경 변수 설정

서비스를 실행하기 전에 다음 환경 변수들을 설정할 수 있습니다. `.env` 파일을 생성하거나 직접 환경 변수를 설정하여 적용합니다.

| 환경 변수 | 설명 | 기본값 | 예시 |
| --- | --- | --- | --- |
| `API_KEY` | **(필수)** API 요청 시 사용할 인증 키입니다. 설정되지 않으면 서버가 500 에러를 반환합니다. | 없음 | `my-secret-key` |
| `ALLOWED_PORTS` | 스캔을 허용할 포트 목록입니다. 단일 포트 및 `-`를 이용한 범위 지정이 가능합니다. 지정한 범위를 벗어난 포트 스캔을 요청하면 거부됩니다. | `1-65535` | `80,443,1000-2000` |
| `REST_API_PORT` | REST API 서버를 노출할 포트 번호입니다. | `8082` | `8082` |

### `.env` 파일 예시

```env
API_KEY=your-secret-key
ALLOWED_PORTS=80,443,8080-8090
REST_API_PORT=8082
```

## 실행 방법

```bash
# 종속성 다운로드
go mod tidy

# 서버 실행 (자동으로 .env 참조)
go run main.go
```

서버는 기본적으로 `:8082` 포트에서 실행됩니다.

## API 사용 예시

### 1. Nmap 스캔 요청 (`POST /api/v1/scan`)

- **Header**: `X-API-Key: {API_KEY}`
- **Content-Type**: `application/json`

**요청 (Request Body):**
```json
{
  "targets": ["127.0.0.1", "scanme.nmap.org"],
  "ports": ["80", "443"]
}
```

**응답 (Response):** 
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<?xml-stylesheet href="file:///usr/bin/../share/nmap/nmap.xsl" type="text/xsl"?>
<nmaprun scanner="nmap" args="nmap -p 80,443 127.0.0.1" ...>
  <!-- Nmap scan results in XML -->
</nmaprun>
```

### 2. Swagger 문서 확인 (`GET /docs`)

브라우저에서 `http://localhost:8082/docs`에 접속하여 Swagger UI를 통해 API 문서를 확인하고 직접 테스트해 볼 수 있습니다.

## 에러 응답 코드

- `400 Bad Request`: 요청 JSON 형식이 잘못되었거나, 타겟이 없거나, `ALLOWED_PORTS`에 허용되지 않은 포트 스캔을 시도한 경우
- `401 Unauthorized`: API 요청 헤더에 `X-API-Key`가 없는 경우
- `403 Forbidden`: 제공된 API 키가 서버의 `API_KEY`와 일치하지 않는 경우
- `499 Client Closed Request`: 클라이언트가 요청을 취소한 경우
- `500 Internal Server Error`: 서버에 API 키가 설정되지 않았거나 Nmap 실행 중 오류가 발생한 경우
- `504 Gateway Timeout`: Nmap 스캔이 5분 이상 소요되어 타임아웃된 경우
