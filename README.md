# ✨ Golang Seeder CLI - Help you seed data easily.
## 🔧 Installation

```sh
go install github.com/komgrip/seeder
```

It is always installed in `~/go/bin`

<br/>

## 📝 Get Started
### 🏗️ **Create seed file**

```sh
seeder create -dir example/seeds/masterdata example_table1
```

```go
//คำสั่งในการสร้าง seed file
seeder create 

//กำหนด directory path ที่ต้องการ ในกรณีที่มี directory path อยู่แล้ว
//จะสร้าง seed file ในนั้น แต่ถ้ายัง command จะสร้าง directory path และ seed file
-dir example/seeds/masterdata

//ชื่อตารางที่เราต้องการจะทำการ seed data
example_table1
```
seed file จะถูก gen ชื่อเป็น `seed_{ชื่อตารางที่เราใส่เข้ามา}_table.sql`

ในกรณีตามตัวอย่างข้างต้นชื่อ seed file จะเป็น `seed_example_table1_table.sql`

<br/>

### 🏗️ **Create multiple seed files**
```sh
seeder create -dir example/seeds/masterdata example_table1 example_table2
```

<br/>

### 📥 **Seed data all files**
```sh
seeder seed -database 'postgres://username:password@localhost:5432/example-db?sslmode=disable' -path databases/seeds/masterdata
```
```go
//คำสั่งใน seed data
seeder seed

//database ที่ต้องการสำหรับ insert ข้อมูล
-database 'postgres://username:password@localhost:5432/example-db?sslmode=disable'
```
<br/>

### 📥 **Seed data specific file**
```sh
seeder seed -database 'postgres://username:password@localhost:5432/example-db?sslmode=disable' -path databases/seeds/masterdata seed_example_table1_table.sql
```
```go
//file ที่ต้องการจะ seed
seed_example_table1_table.sql
```