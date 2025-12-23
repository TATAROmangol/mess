mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc admin user add myminio avatar-backend avatarback
mc admin policy create myminio avatar-policy avatar-backend-policy.json
mc admin policy attach myminio avatar-policy --user avatar-backend

