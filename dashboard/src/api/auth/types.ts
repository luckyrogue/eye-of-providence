type AuthResponse = {
    token: string;
    user_id: string;
    display_name?: string;
    team_id?: string | null;
  };
  
type AuthConfig = { invite_only: boolean; is_first_user: boolean };

export type { AuthResponse, AuthConfig };