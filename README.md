#🍲 RecipeShare 
An AI-powered user-centric back-end for sharing food recipes in Golang, using GORM, and Gin frameworks.

## ✨ Features  

- 👤 **User Management**  
  - User signup/login with secure password hashing  
  - JWT authentication  
  - Email verification system  

- 📖 **Recipe Management**  
  - Create, update, delete, and view recipes  
  - Add ingredients and preparation steps  
  - Recipe versioning (track updates over time)  

- ⭐ **Engagement**  
  - Rate recipes  
  - Leave comments  
  - Add recipes to favorites  
  - Tagging and categorization  

- 📊 **Analytics & Admin Tools**  
  - Recipe popularity tracking  
  - User activity logs  
  - Admin-level recipe and user management  

- 🥗 **Nutrition Features**  
  - Basic nutrition estimation per recipe  

---

## 🛠️ Tech Stack  

**Backend:**  
- [Go (Golang)](https://golang.org/) with [Gin](https://gin-gonic.com/)  
- [GORM](https://gorm.io/) ORM  
- [MySQL](https://www.mysql.com/) Database  
- JWT Authentication  

**Frontend (in progress):**  
- Angular  
- TailwindCSS  

---

## 📂 Project Structure  
```
├───.vscode
├───controller
├───docs
├───internal
├───messages
├───middleware
├───migrate
├───model
├───routes
├───service
├───token
├───uploads
└───utils
```
## ⚡ Getting Started  

### Prerequisites  
- Go 1.22+  
- MySQL  
- Git  

### Setup  

1. Clone the repository  
```bash
   git clone https://github.com/your-username/recipeshare.git
   cd recipeshare
```
2. Configure the database connection in internal/database.go

3. Run database migrations
```bsh

go run main.go migrate

```

4. Start the server
```bash

go run main.go

```

5. Visit APIs at:
```

http://localhost:3000

```
