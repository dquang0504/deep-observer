# Deep-Observer Learning Journal

Đây là nơi lưu trữ các kiến thức kỹ thuật và giải đáp thắc mắc trong quá trình phát triển dự án Deep-Observer.

## 1. Cơ chế Web & JSON (Gin Framework)

### `c.ShouldBindJSON` vs `json.Unmarshal`
- **Câu hỏi**: `c.ShouldBindJSON` có giống `json.Unmarshal` không?
- **Trả lời**: 
    - `json.Unmarshal`: Chỉ parse bytes JSON thành Struct.
    - `c.ShouldBindJSON`: Là bản nâng cấp cho Web. Nó tự động đọc Request Body, parse JSON, và **kiểm tra tính hợp lệ (Validation)** dựa trên các tag như `binding:"required"`.

## 2. Quản lý Cơ sở dữ liệu (Database)

### Tại sao cần `defer rows.Close()`?
- **Câu hỏi**: Có bị memory leak nếu không đóng rows không?
- **Trả lời**: **CÓ**. Nếu không đóng, kết nối (connection) sẽ không được trả lại cho "Connection Pool". Khi pool hết kết nối, server sẽ **treo (hang)** vì các request mới phải đợi kết nối trống mãi mãi.

### Connection Pool Hanging (Cơ chế chi tiết)
- **Tình huống**: Pool có giới hạn (ví dụ 20 kết nối). Mỗi lần quên `Close()`, bạn làm mất 1 kết nối.
- **Hậu quả**: Khi hết 20 kết nối, request thứ 21 sẽ đứng đợi vô hạn. Restart container chỉ là giải pháp tạm thời vì nó xóa sạch các kết nối cũ, nhưng lỗi code sẽ sớm làm cạn kiệt pool lần nữa.

### `QueryRow` vs `Query`
- **Câu hỏi**: `QueryRow` là lệnh Select đúng không? `Scan` là gì?
- **Trả lời**:
    - `QueryRow`: Dùng khi truy vấn trả về **duy nhất 1 dòng**. Nó tự động xử lý Close cho bạn.
    - `Scan`: Là lệnh "đổ" dữ liệu từ các cột trong Database vào các biến trong Go thông qua địa chỉ bộ nhớ (`&`).

## 3. Ngôn ngữ Go (Golang)

### Struct Value vs Struct Pointer
- **Câu hỏi**: Tại sao dùng `var s model.Service` mà không phải `*model.Service`? Dùng pointer chẳng phải tránh việc copy dữ liệu sao?
- **Trả lời**: 
    - Trong Go, dùng `var s model.Service` cấp phát bộ nhớ ngay lập tức. Khi cần tránh copy, ta truyền địa chỉ `&s` vào hàm.
    - Nếu dùng `var s *model.Service` mà không khởi tạo (`new`), biến sẽ là `nil`. Khi `Scan` dữ liệu vào một biến `nil` sẽ gây ra lỗi Crash (Panic) chương trình.

### Xử lý giá trị Null từ Database
- **Câu hỏi**: Tại sao cần biến trung gian `*string` cho các cột nullable?
- **Trả lời**: Kiểu `string` trong Go không thể nhận giá trị `NULL` từ DB (nó sẽ báo lỗi). Ta dùng `*string` (con trỏ) vì con trỏ có thể mang giá trị `nil`, tương ứng hoàn hảo với `NULL` trong SQL. Sau khi Scan, ta kiểm tra `!= nil` trước khi thực hiện gán vào Struct chính.
