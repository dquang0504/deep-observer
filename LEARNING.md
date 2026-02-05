### Kinh nghiệm xương máu: Lỗi "No Data" dù Test Connection OK
- **Triệu chứng**: Cấu hình file `datasources.yaml`, vào giao diện bấm "Save & Test" báo xanh (OK), nhưng ra Dashboard lại báo "No Data" hoặc "Missing database".
- **Nguyên nhân**: Grafana (các bản mới) ưu tiên đọc tên database nằm trong `jsonData` thay vì ở ngoài root. Nếu thiếu, nó kết nối được tới Server nhưng không biết query vào Database nào.
- **Giải pháp**: Luôn khai báo tên database ở cả hai nơi:
    ```yaml
    database: my_db_name       # Cho plugin cũ
    jsonData:
      database: my_db_name     # Quan trọng cho plugin mới
    ```

## 11. OpenTelemetry & Go (Deep Dive)

### Khái niệm cơ bản
- **OTEL (OpenTelemetry)**: Tiêu chuẩn mở để thu thập dữ liệu quan sát (Metrics, Logs, Traces).
- **Core (`/otel`)**: API trừu tượng (Giao diện).
- **SDK (`/otel/sdk`)**: Bộ máy thực thi (Xử lý dữ liệu).
- **Exporters (`/otel/exporters/...`)**: Người vận chuyển (Gửi dữ liệu đi).

### Pattern `initTelemetry` (Hàm trả về hàm)
- **Tên gọi**: Closure hoặc Higher-Order Function.
- **Mục đích**: Đóng gói logic khởi tạo (Init) và dọn dẹp (Cleanup) vào cùng một nơi.
- **Ví dụ luồng chạy**:
    ```mermaid
    sequenceDiagram
        Main->>initTelemetry: Gọi hàm khởi tạo
        initTelemetry->>OTEL Collector: Kết nối HTTP/gRPC
        initTelemetry-->>Main: Trả về hàm 'shutdown'
        Main->>Main: Chạy logic chính (Gửi Event...)
        Main->>shutdown: Gọi hàm 'shutdown' (dùng defer)
        shutdown->>OTEL Collector: Gửi nốt dữ liệu còn sót & Ngắt kết nối
    ```

### Luồng xử lý Metric & Trace trong code
```mermaid
graph TD
    A[Start App] --> B[Init Telemetry]
    B --> C[Tạo Tracer & Meter]
    C --> D[Tạo Span 'main_execution']
    D --> E{Gửi Event SDK}
    E -- Success --> F[Ghi Status OK vào Span]
    E -- Error --> G[Ghi Error vào Span]
    F & G --> H[Kết thúc Span]
    H --> I[Shutdown (Gửi dữ liệu về Collector)]

## 12. Cấu hình Observability Chuyên sâu

### Loki Configuration (`loki-config.yml`)
- **Storage**: Sử dụng `filesystem` thay vì S3/GCS. Đây là lựa chọn tốt nhất cho môi trường local/dev. Dữ liệu được lưu tại `/tmp/loki/chunks`.
- **TSDB vs BoltDB**:
    - `tsdb` (Time Series DB): Phiên bản mới, hiệu năng cao hơn, nén tốt hơn.
    - `boltdb`: Phiên bản cũ, dựa trên key-value store.
- **Compaction**: Quá trình gộp các mảnh log nhỏ thành mảnh lớn và xóa log cũ (Retention Policy).

### Tempo Configuration (`tempo.yml`)
- **Kiến trúc Monolith**: Chạy tất cả các thành phần trong 1 binary cho đơn giản.
- **Distributor**: Cổng đón nhận dữ liệu.
- **Ingester**: Bộ nhớ đệm (Buffer). Dữ liệu nằm ở đây trước khi được ghi xuống ổ cứng.
- **WAL (Write Ahead Log)**: Cơ chế đảm bảo an toàn dữ liệu. Ghi log thao tác ra đĩa trước khi thực hiện để recover nếu sập nguồn.
- **Metrics Generator**: Tính năng cực mạnh. Tempo tự động phân tích trace và tạo ra metrics (số lượng request, độ trễ) gửi sang Prometheus. Giúp ta vẽ biểu đồ RED (Rate, Errors, Duration) mà không cần code thêm metric trong App.

### OpenTelemetry Collector (`collector.yaml`)
- **Vai trò**: "Người phiên dịch" và "Router".
- **Pipeline**: `Receiver` -> `Processor` -> `Exporter`.
- **Tại sao dùng `otlphttp/loki`?**:
    - Exporter `loki` cũ đã bị loại bỏ.
    - Grafana Loki hiện đại hỗ trợ nhận log qua chuẩn OTLP.
    - Cấu hình này giúp Collector nói chuyện với Loki một cách chuẩn mực.
```
