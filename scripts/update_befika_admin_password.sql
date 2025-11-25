-- Script to update password for befika_admin user
-- Password hash: $2a$12$Yap6vkUgav2TUzZjZPTfPehKjXiQWwa2tYHFP8vim0LKBgsLbd7aO

UPDATE users 
SET password = '$2a$12$Yap6vkUgav2TUzZjZPTfPehKjXiQWwa2tYHFP8vim0LKBgsLbd7aO' 
WHERE username = 'befika_admin';

-- Verify the update
SELECT 
    username, 
    email, 
    is_admin, 
    user_type,
    CASE 
        WHEN password = '$2a$12$Yap6vkUgav2TUzZjZPTfPehKjXiQWwa2tYHFP8vim0LKBgsLbd7aO' 
        THEN 'Password updated successfully' 
        ELSE 'Password NOT updated' 
    END as password_status
FROM users 
WHERE username = 'befika_admin';

