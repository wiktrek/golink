CREATE TABLE IF NOT EXISTS `User` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(96) NOT NULL UNIQUE,
    `password` VARCHAR(128) NOT NULL,
    `email` VARCHAR(128) NOT NULL UNIQUE,
    `createdAt` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS `Short` (
    `slug` VARCHAR(256) PRIMARY KEY,
    `url` TEXT NOT NULL,
    `ownerId` INT,
    `createdAt` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `expiresAt` TIMESTAMP,
    `isActive` BOOLEAN DEFAULT TRUE;
    `clicks` INT DEFAULT 0,
    FOREIGN KEY (`ownerId`) REFERENCES `User`(`id`) ON DELETE SET NULL
);

CREATE INDEX idx_user_email ON `User`(`email`);
CREATE INDEX idx_short_ownerId ON `Short`(`ownerId`);