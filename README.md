
# Golang Nigeria Job Board 🇳🇬  

A fast and lightweight job board built with **Golang**, **HTMX**, **Alpine.js**, and **Templ**. Find the latest Golang job openings in Nigeria with instant filtering and a SPA-like experience.  

## Features  
- 🚀 **Fast Loading:** Fetches the latest 10 jobs first, then loads the rest in the background.  
- 🗂 **Filter Jobs:** Easily filter by remote work, location, and more using Templ.  
- ⚡ **Cache for Speed:** Jobs are cached in Redis to reduce database queries.  
- 🎨 **Modern UI:** Built with **Hyper UI** for a clean and responsive design.  

## Tech Stack  
- **Backend:** Golang 
- **Frontend:** HTMX, Alpine.js, Templ  
- **Database:** PostgreSQL  
- **Cache:** Redis  

## Installation  
```sh
git clone https://github.com/Cprime50/GoCafeNG.git
cd GoCafeNG.git  
go run main.go  
