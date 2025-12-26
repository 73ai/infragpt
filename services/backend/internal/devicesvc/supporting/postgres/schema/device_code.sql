CREATE TABLE device_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_code TEXT UNIQUE NOT NULL,
    user_code VARCHAR(10) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    organization_id UUID,
    user_id UUID,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_device_codes_user_code ON device_codes (user_code);
CREATE INDEX idx_device_codes_device_code ON device_codes (device_code);
CREATE INDEX idx_device_codes_expires_at ON device_codes (expires_at);
