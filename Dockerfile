FROM curlimages/curl:8.22.0

# 基礎鏡像已含 curl 與 CA 憑證；郵件備援流程仍可呼叫外部服務。
COPY dist/app /usr/local/bin/app

ENTRYPOINT ["/usr/local/bin/app"]
