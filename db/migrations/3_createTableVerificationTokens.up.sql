CREATE TABLE verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, 
    token VARCHAR(256) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
