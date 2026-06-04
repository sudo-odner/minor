import React, { useState, useEffect } from 'react';
import { getServerMembers, Member } from '../../api/members';

interface MemberListProps {
  serverId: string | null;
  onlineUsers?: Record<string, boolean>;
}

const MemberList: React.FC<MemberListProps> = ({ serverId, onlineUsers = {} }) => {
  const [members, setMembers] = useState<Member[]>([]);

  useEffect(() => {
    if (!serverId) return;
  
    const fetchMembers = async () => {
      try {
        const data = await getServerMembers(serverId);
        
        if (Array.isArray(data)) {
          setMembers(data);
        } else {
          console.error("Бэкенд вернул не массив участников:", data);
          setMembers([]);
        }
      } catch (err) {
        console.error("Ошибка при загрузке участников:", err);
        setMembers([]);
      }
    };
  
    fetchMembers();
  }, [serverId]);

  if (!serverId) return null;

  // Apply real-time presence overrides from parent
  const membersWithPresence = members.map(m => {
    if (m.user_id in onlineUsers) {
      return {
        ...m,
        status: onlineUsers[m.user_id] ? 'USER_STATUS_ONLINE' : 'USER_STATUS_OFFLINE',
      };
    }
    return m;
  });

  // Группируем участников на "В сети" и "Не в сети"
  const onlineMembers = membersWithPresence.filter(m => m.status === 'USER_STATUS_ONLINE');
  const offlineMembers = membersWithPresence.filter(m => m.status !== 'USER_STATUS_ONLINE');

  const renderMemberRow = (member: Member) => {
    const displayName = member.nickname || member.username || 'Unknown';
    const avatarLetter = displayName.charAt(0).toUpperCase() || '?';
    const isOnline = member.status === 'USER_STATUS_ONLINE';
  
  return (
    <div key={member.user_id} className="flex items-center space-x-2 p-1.5 rounded hover:bg-brand-blue-light/50 cursor-pointer group transition-colors">
      <div className="relative">
        <div className="w-8 h-8 bg-brand-blue/10 border border-brand-blue-light rounded-full flex items-center justify-center font-bold text-xs uppercase text-brand-blue transition-colors">
          {avatarLetter}
        </div>
        <div className={`absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-white transition-colors ${
          isOnline ? 'bg-[#23a55a]' : 'bg-gray-300'
        }`} />
      </div>
      <div className="text-sm truncate">
        <p className="font-medium text-gray-600 group-hover:text-brand-blue truncate transition-colors">{displayName}</p>
        {member.nickname && <p className="text-[10px] text-gray-400 truncate font-medium">@{member.username}</p>}
      </div>
    </div>
  );
};

return (
  <div className="w-60 bg-white border-l border-brand-blue-light p-4 flex flex-col space-y-4 overflow-y-auto no-scrollbar hidden md:flex transition-colors duration-200">
    {/* Секция "В сети" */}
    {onlineMembers.length > 0 && (
      <div className="space-y-1">
        <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest px-1">
          В сети — {onlineMembers.length}
        </p>
        {onlineMembers.map(member => renderMemberRow(member))}
      </div>
    )}

    {/* Секция "Не в сети" */}
    {offlineMembers.length > 0 && (
      <div className="space-y-1">
        <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest px-1 mt-2">
          Не в сети — {offlineMembers.length}
        </p>
        {offlineMembers.map(member => renderMemberRow(member))}
      </div>
    )}

    {members.length === 0 && (
      <div className="text-xs text-gray-400 italic text-center py-8">Участников пока нет</div>
    )}
  </div>
);
};

export default MemberList;