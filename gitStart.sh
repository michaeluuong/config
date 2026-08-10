go env -w GOPRIVATE=github.com/michaeluuong/*
git init
git add .
git commit -m ""
git remote add origin https://github.com/michaeluuong/config.git
git push -u origin main
