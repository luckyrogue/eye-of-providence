// Тела запросов, отправляемых на /v1/auth/*

type RegisterReq = {
  email: string;
  password: string;
  display_name: string;
  invite_code?: string;
};

type LoginReq = {
  email: string;
  password: string;
};

export type { RegisterReq, LoginReq };
