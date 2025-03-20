#!/bin/bash

echo "Creating .env file - Interactive Setup"
echo "------------------------------------"
echo "Press Enter to use default value (shown in brackets)"
echo ""

# Database settings
read -p "PostgreSQL username [badjokes]: " db_user
db_user=${db_user:-badjokes}

read -p "PostgreSQL password [qwerty123]: " db_password
db_password=${db_password:-qwerty123}

read -p "PostgreSQL database name [bad_jokes]: " db_name
db_name=${db_name:-bad_jokes}

# JWT settings
read -p "JWT secret key [jwt_secret_value]: " jwt_secret
jwt_secret=${jwt_secret:-jwt_secret_value}

# URLs
read -p "Backend URL [http://localhost:9999]: " backend_url
backend_url=${backend_url:-http://localhost:9999}

read -p "Base OAuth URL [http://localhost:9999]: " base_oauth_url
base_oauth_url=${base_oauth_url:-http://localhost:9999}

read -p "Callback OAuth URL [http://localhost:5173]: " callback_url
callback_url=${callback_url:-http://localhost:5173}

# OAuth credentials
read -p "Google Client ID: " google_id
read -p "Google Client Secret: " google_secret

read -p "GitHub Client ID: " github_id
read -p "GitHub Client Secret: " github_secret

# Create .env file
cat > .env << EOF
POSTGRES_USER=${db_user}
POSTGRES_PASSWORD=${db_password}
POSTGRES_DB=${db_name}
JWT_SECRET_KEY=${jwt_secret}
BACKEND_URL=${backend_url}
GOOGLE_CLIENT_ID=${google_id}
GOOGLE_CLIENT_SECRET=${google_secret}
GITHUB_CLIENT_ID=${github_id}
GITHUB_CLIENT_SECRET=${github_secret}
BASE_OAUTH_URL=${base_oauth_url}
CALLBACK_OAUTH_URL=${callback_url}
EOF

echo ""
echo ".env file has been created successfully!"