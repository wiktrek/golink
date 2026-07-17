CREATE TABLE IF NOT EXISTS `User` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(96) NOT NULL UNIQUE,
    `password` VARCHAR(255) NOT NULL,
    `email` VARCHAR(128) NOT NULL UNIQUE,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS `Short` (
    `slug` VARCHAR(64) PRIMARY KEY,
    `url` TEXT NOT NULL,
    `ownerId` INT NOT NULL,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expiresAt` TIMESTAMP NULL,
    `isActive` BOOLEAN NOT NULL DEFAULT TRUE,
    `clicks` INT NOT NULL DEFAULT 0,
    CONSTRAINT `short_owner_fk` FOREIGN KEY (`ownerId`) REFERENCES `User`(`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `Session` (
    `tokenHash` VARCHAR(64) PRIMARY KEY,
    `userId` INT NOT NULL,
    `expiresAt` TIMESTAMP NOT NULL,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `session_user_fk` FOREIGN KEY (`userId`) REFERENCES `User`(`id`) ON DELETE CASCADE
);

CREATE INDEX `idx_short_owner_id` ON `Short` (`ownerId`);
CREATE INDEX `idx_short_expires_at` ON `Short` (`expiresAt`);
CREATE INDEX `idx_session_expires_at` ON `Session` (`expiresAt`);
