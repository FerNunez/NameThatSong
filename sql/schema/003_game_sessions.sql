-- +goose Up
-- Create game sessions table for tracking game history
CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_mode VARCHAR(50) NOT NULL,
    total_questions INTEGER NOT NULL,
    correct_answers INTEGER NOT NULL,
    total_score INTEGER NOT NULL,
    duration_seconds INTEGER NOT NULL,
    tracks_played TEXT[], -- Array of Spotify track IDs
    completed_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for game sessions
CREATE INDEX idx_game_sessions_user_id ON game_sessions(user_id);
CREATE INDEX idx_game_sessions_completed_at ON game_sessions(completed_at);
CREATE INDEX idx_game_sessions_game_mode ON game_sessions(game_mode);

-- +goose Down
-- Drop tables in reverse order
DROP TABLE game_sessions CASCADE;
