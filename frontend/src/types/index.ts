export interface TelegramUser {
  id: number
  username?: string
  first_name?: string
  last_name?: string
  photo_url?: string
}

export interface Player {
  id: number
  room_id: string
  user_id: number
  username?: string
  first_name?: string
  team?: string
  score: number
  is_host: boolean
  joined_at: string
}

export interface Room {
  id: string
  status: 'lobby' | 'playing' | 'finished'
  current_round: number
  current_explainer_id?: number
  round_end_at?: string
  category: string
  num_teams: number
  team_names: string[]
  created_at: string
}

export interface Word {
  id: number
  word: string
}

export interface GameState {
  room: Room | null
  players: Player[]
  currentWord: Word | null
  secondsLeft: number
  teamScores: Record<string, number>
}

// WebSocket messages
export type WSMessageType =
  | 'player_joined'
  | 'player_left'
  | 'team_changed'
  | 'game_started'
  | 'new_word'
  | 'word_result'
  | 'timer'
  | 'round_end'
  | 'game_end'
  | 'error'
  | 'room_state'
  | 'score_update'
  | 'swipe'

export interface WSMessage {
  type: WSMessageType
  payload?: unknown
}

export interface PlayerJoinedPayload {
  player: Player
}

export interface PlayerLeftPayload {
  user_id: number
}

export interface TeamChangedPayload {
  user_id: number
  team: string
}

export interface GameStartedPayload {
  explainer_id: number
  round_end_at: number
}

export interface NewWordPayload {
  word_id: number
  word: string
}

export interface WordResultPayload {
  word_id: number
  word: string
  guessed: boolean
}

export interface TimerPayload {
  seconds_left: number
}

export interface RoundEndPayload {
  round: number
  team_scores: Record<string, number>
  next_explainer: number
}

export interface GameEndPayload {
  winner: string
  team_scores: Record<string, number>
}

export interface GameStats {
  room_id: string
  team_scores: Record<string, number>
  players: PlayerStats[]
  rounds: RoundStats[]
}

export interface PlayerStats {
  user_id: number
  first_name: string
  team: string
  score: number
  words_guessed: number
  words_missed: number
}

export interface RoundStats {
  round_num: number
  explainer_id: number
  words_guessed: number
  words_missed: number
}
