# WebSocket untuk Streaming Data Gambar ESP

Implementasi WebSocket ini memungkinkan frontend untuk menerima update gambar secara realtime dari perangkat ESP.

## Format Data MQTT

```json
{
  "X-API-KEY": "your-api-key-here", 
  "X-MAC-ADDRESS": "esp-mac-address",
  "image": "base64-encoded-image-data"
}
```

## Endpoint WebSocket

1. **Stream Gambar dari ESP Tertentu**
   - URL: `/ws/device?esp_hmac=MAC_ADDRESS`
   - Format Data:
     ```json
     {
       "esp_hmac": "00:11:22:33:44:55",
       "image_data": "data:image/jpeg;base64,...",
       "timestamp": 1617345600000
     }
     ```

2. **Stream Semua Perangkat ESP (Admin)**
   - URL: `/ws/devices/all`
   - Memerlukan autentikasi admin
   - Format Data (array objek):
     ```json
     [
       {
         "esp_hmac": "00:11:22:33:44:55",
         "image_data": "data:image/jpeg;base64,...",
         "timestamp": 1617345600000
       },
       {
         "esp_hmac": "AA:BB:CC:DD:EE:FF",
         "image_data": "data:image/jpeg;base64,...",
         "timestamp": 1617345600000
       }
     ]
     ```

## REST API untuk Monitoring ESP

1. **Mendapatkan Daftar Semua ESP (Admin)**
   - Method: `GET`
   - URL: `/v1/devices`
   - Memerlukan autentikasi admin
   - Response:
     ```json
     {
       "status": "success",
       "count": 2,
       "devices": [
         {
           "esp_hmac": "00:11:22:33:44:55",
           "image_data": "data:image/jpeg;base64,...",
           "timestamp": 1617345600000
         },
         {
           "esp_hmac": "AA:BB:CC:DD:EE:FF",
           "image_data": "data:image/jpeg;base64,...",
           "timestamp": 1617345600000
         }
       ]
     }
     ```

2. **Mendapatkan Gambar Terbaru dari ESP Tertentu (Admin)**
   - Method: `GET`
   - URL: `/v1/devices/:esp_hmac`
   - Memerlukan autentikasi admin
   - Response:
     ```json
     {
       "esp_hmac": "00:11:22:33:44:55",
       "image_data": "data:image/jpeg;base64,...",
       "timestamp": 1617345600000
     }
     ```

## Cara Kerja

1. ESP mengirim data gambar ke MQTT broker
2. Backend menerima data dari MQTT dan menyimpannya dalam memori
3. Backend meneruskan data ke client yang terhubung via WebSocket
4. Admin dapat melihat semua ESP terhubung untuk kemudian dikaitkan dengan slot parkir

## Keamanan

- API Key digunakan untuk memvalidasi pesan MQTT
- Endpoint admin dilindungi dengan middleware autentikasi
- WebSocket untuk "all devices" dilindungi dengan middleware admin 