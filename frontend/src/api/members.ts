import { api } from './axios';

export interface Member {
  userId: string;
  serverId: string;
  nickname?: string;
  username: string;      // Обогащается на бэкенде из User Service
  avatarUrl?: string;
  status?: string;       // Обогащается на бэкенде из Presence Service
  roles: string[];       // Массив ID ролей участника
}

// 1. Вступить на сервер (AddMember)
export const joinServer = async (serverId: string): Promise<Member> => {
  const response = await api.post(`/servers/${serverId}/members`);
  return response.data;
};

// 2. Получить список участников сервера (GetServerMembers)
export const getServerMembers = async (serverId: string): Promise<Member[]> => {
  const response = await api.get(`/servers/${serverId}/members`);
  return Array.isArray(response.data) ? response.data : [];
};

// 3. Изменить никнейм на сервере (UpdateNickname)
export const updateNickname = async (serverId: string, userId: string, nickname: string): Promise<void> => {
  await api.patch(`/servers/${serverId}/members/${userId}/nickname`, { nickname });
};

// 4. Назначить роль участнику (AddRoleToMember)
export const assignRoleToMember = async (serverId: string, userId: string, roleId: string): Promise<void> => {
  await api.patch(`/servers/${serverId}/members/${userId}/roles/${roleId}`);
};

// 5. Забрать роль у участника (RemoveRoleFromMember)
export const removeRoleFromMember = async (serverId: string, userId: string, roleId: string): Promise<void> => {
  await api.delete(`/servers/${serverId}/members/${userId}/roles/${roleId}`);
};

// 6. Выгнать участника с сервера (RemoveMember / Kick)
export const kickMember = async (serverId: string, userId: string): Promise<void> => {
  await api.delete(`/servers/${serverId}/members/${userId}`);
};