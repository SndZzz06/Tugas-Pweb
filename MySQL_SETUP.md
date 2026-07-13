# NextState Backend - MySQL Setup Guide

## 🚨 Current Problem
```
Database initialization failed: failed to connect to database
dial tcp 127.0.0.1:3306: connectex: No connection could be made
```

**Artinya:** MySQL belum berjalan atau database belum dibuat.

---

## ✅ Solusi: Setup MySQL Step by Step

### Step 1: Nyalakan MySQL (PENTING!)

#### Metode A: Menggunakan XAMPP Control Panel
1. Buka folder: `C:\xampp`
2. Jalankan: `xampp-control.exe`
3. Cari **MySQL** dalam daftar
4. Klik tombol **"Start"** di sebelah MySQL
5. Tunggu status berubah menjadi **"Running"** (text hijau)

#### Metode B: Menggunakan Command Line
```powershell
net start MySQL80
```

#### Verifikasi MySQL sudah jalan:
```powershell
mysql -u root -e "SELECT 1"
```
Jika berhasil, akan tampil: `1`

---

### Step 2: Setup Database

#### Metode A: Menggunakan Batch File (Paling Mudah)
1. Buka Windows Explorer
2. Navigate ke: `C:\xampp\htdocs\nextState-main\BE\golang\`
3. **Double-click**: `setup_db.bat`
4. Tunggu sampai muncul "SUCCESS! Database setup complete!"

#### Metode B: Menggunakan Command Line
```powershell
cd C:\xampp\htdocs\nextState-main\BE\golang
mysql -u root < schema.sql
```

#### Verifikasi Database berhasil dibuat:
```powershell
mysql -u root -e "USE nextstate_db; SHOW TABLES;"
```

Harus tampil:
```
Tables_in_nextstate_db
projects
task_user
tasks
users
```

---

## 🎮 Menjalankan Server

### Metode 1: Double-click File Batch (PALING MUDAH)
1. Buka folder: `C:\xampp\htdocs\nextState-main\BE\golang\`
2. **Double-click**: `run.bat`
3. Server akan start di port 8080

### Metode 2: Command Line
```powershell
cd C:\xampp\htdocs\nextState-main\BE\golang
go run .
```

### Expected Output (Jika sukses):
```
Database connection established successfully.
NextState Go server running at http://localhost:8080...
```

---

## 🌐 Akses Aplikasi

Setelah server running, buka browser:

- **Login Page**: http://localhost:8080/login
- **Register Page**: http://localhost:8080/register
- **Dashboard**: http://localhost:8080/ (perlu login dulu)

---

## 🔧 Troubleshooting

### Error: "MySQL is not running"
```
ERROR: MySQL is not running or not accessible!
```
**Solusi:**
1. Buka XAMPP Control Panel
2. Klik "Start" untuk MySQL
3. Tunggu sampai status berubah menjadi "Running"
4. Coba lagi

### Error: "Unknown database"
```
ERROR 1049: Unknown database 'nextstate_db'
```
**Solusi:**
```powershell
mysql -u root < schema.sql
```

### Error: "Access denied for user 'root'"
Jika MySQL root punya password, edit `.env`:
```
DB_USERNAME=root
DB_PASSWORD=your_password_here
```

### Port 8080 already in use
Jika port 8080 sudah dipakai, edit `.env`:
```
PORT=9000
```
Kemudian akses: http://localhost:9000

---

## 📋 Checklist

- [ ] MySQL running (XAMPP Control Panel)
- [ ] Database created (`setup_db.bat` sudah dijalankan)
- [ ] `.env` file dikonfigurasi dengan benar
- [ ] Run `go run .`
- [ ] Server berjalan di http://localhost:8080

---

## 📝 Database Schema

Tabel yang dibuat:

1. **users** - User accounts dengan role (Admin/Member)
2. **projects** - Project list
3. **tasks** - Tasks dalam setiap project
4. **task_user** - Relasi assign task ke users

---

**Catatan:** Jika sudah dijalankan sekali, untuk menjalankan lagi cukup:
```powershell
go run .
```

Tidak perlu setup database lagi (kecuali drop/delete database).
