package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Email    EmailConfig    `yaml:"email"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

type EmailConfig struct {
	SMTPHost   string `yaml:"smtp_host"`
	SMTPPort   int    `yaml:"smtp_port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Sender     string `yaml:"sender"`
	AdminEmail string `yaml:"admin_email"`
}

type JWTConfig struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Name, c.User, c.Password, c.SSLMode,
	)
}

func (s *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
