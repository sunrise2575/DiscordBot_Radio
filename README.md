# Moooclub Radio

Golang으로 제작한 24시간 디스코드 음악 스트리밍 봇

- 자기 서버의 폴더를 재귀적으로 읽어서 무작위로 음악 재생
- 여러 디스코드 서버에서 실행 가능
- `::` 명령으로 음악 넘기기 가능

## 실행방법

1. `config.json` 을 작성한다 (`config.example.json` 참고)
   - `discord_token`: 디스코드 봇 토큰
   - `folder_path`: 음악 파일들이 들어가 있는 폴더 (재귀적으로 탐색합니다)

2. 아래 명령어로 실행
   ```
   cd src/
   go run .
   ```

## 특징 
- in-memory sqlite3 사용하여 코드 로직 간소화
- 최소기능주의 설계