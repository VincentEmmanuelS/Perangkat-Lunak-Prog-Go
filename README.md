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
- Client dapat membuat room chat baru.
  - Room chat dapat memiliki batas jumlah client (default tanpa batas).
  - Room chat dapat diberi password (default tanpa password).
  - Room chat dapat memiliki batas client dan password sekaligus.
- Client dapat melihat daftar semua room chat yang tersedia.
- Client dapat bergabung ke room chat yang tersedia.
- Client dapat mengirim dan menerima pesan dari client lain dalam room yang sama.
- Client dapat keluar dari room chat.
- Client dapat berpindah ke room chat lain.

---


## Informasi Tambahan

- Server berjalan secara default di alamat `localhost:8080`.
- Untuk uji coba lokal, client dapat dijalankan di beberapa terminal berbeda.

---

## Collaborator

- Vincent Emmanuel Suwardy
- Michael William Iswadi
- Stanislaus Nathan
- Jensen Hiem

---

## Referensi

- [Go 'net' Documentation](https://pkg.go.dev/net)
- [Go 'Defer, Panic, and Recover' Documentation](https://go.dev/blog/defer-panic-and-recover)
- [Go Documentation by w3schools](https://www.w3schools.com/go/)
- [Go 'net' Documentation by Sling Academy](https://www.slingacademy.com/article/using-the-net-package-for-low-level-network-programming-in-go/)
- Semua slide kuliah
