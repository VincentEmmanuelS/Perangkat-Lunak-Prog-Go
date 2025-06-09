# Chat Room TCP Sederhana (Golang)

Proyek ini merupakan implementasi sederhana dari sistem **chat room berbasis TCP** menggunakan bahasa pemrograman **Go (Golang)**. Aplikasi ini memungkinkan banyak pengguna (client) terhubung ke satu server dan berkomunikasi dalam satu ruang obrolan secara real-time.

Dibuat untuk memenuhi tugas besar mata kuliah **AIF233131 Pemrograman Go**.

---

## Fitur

- Server dapat menerima banyak koneksi client secara bersamaan.
- Setiap client:
  - Dapat memilih username saat terhubung.
  - Menerima pesan siaran dari client lain yang terhubung.
  - Melihat notifikasi ketika client lain bergabung atau keluar dari chat.
- Sistem menggunakan **goroutine** untuk menangani koneksi secara paralel.
- Komunikasi dilakukan sepenuhnya melalui protokol **TCP**.

---


## Informasi Tambahan

- Server berjalan secara default di alamat `localhost:8080`.
- Untuk uji coba lokal, client dapat dijalankan di beberapa terminal berbeda.
- Proyek ini bersifat **open-source** dan dapat dikembangkan lebih lanjut untuk menambahkan fitur seperti:
  - Private chat (whisper)
  - Otentikasi pengguna
  - Penyimpanan histori chat

---

## Collaborator

- Vincent Emmanuel Suwardy / 6182201067
- Stanislaus Nathan / 6182201092
- Michael William Iswadi / 6182201019

---

## Referensi

- [Go 'net' Documentation](https://pkg.go.dev/net)
- [Go 'Defer, Panic, and Recover' Documentation](https://go.dev/blog/defer-panic-and-recover)
- [Go Documentation by w3schools](https://www.w3schools.com/go/)
- [Go 'net' Documentation by Sling Academy](https://www.slingacademy.com/article/using-the-net-package-for-low-level-network-programming-in-go/)
- Semua slide kuliah
