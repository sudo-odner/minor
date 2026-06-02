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
      <div key={member.user_id} className="flex items-center space-x-2 p-1.5 rounded hover:bg-[#e3e5e8] dark:hover:bg-[#35373c] cursor-pointer group transition-colors">
        <div className="relative">
          <div className="w-8 h-8 bg-gray-400 dark:bg-gray-600 rounded-full flex items-center justify-center font-bold text-xs uppercase text-white dark:text-gray-200">
            {avatarLetter}
          </div>
          <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-[#f2f3f5] dark:border-[#2b2d31] transition-colors ${
            isOnline ? 'bg-[#23a55a]' : 'bg-gray-500'
          }`} />
        </div>
        <div className="text-sm truncate">
          <p className="font-medium text-[#4f5660] dark:text-gray-300 group-hover:text-[#060607] dark:group-hover:text-white truncate transition-colors">{displayName}</p>
          {member.nickname && <p className="text-[10px] text-[#4f5660] dark:text-gray-500 truncate">@{member.username}</p>}
        </div>
      </div>
    );
  };

  return (
    <div className="w-60 bg-[#f2f3f5] dark:bg-[#2b2d31] border-l border-[#e3e5e8] dark:border-[#1e1f22] p-4 flex flex-col space-y-4 overflow-y-auto no-scrollbar hidden md:flex transition-colors duration-200">
      {/* Секция "В сети" */}
      {onlineMembers.length > 0 && (
        <div className="space-y-1">
          <p className="text-[10px] font-bold text-[#4f5660] dark:text-gray-400 uppercase tracking-wider">
            В сети — {onlineMembers.length}
          </p>
          {onlineMembers.map(member => renderMemberRow(member))}
        </div>
      )}

      {/* Секция "Не в сети" */}
      {offlineMembers.length > 0 && (
        <div className="space-y-1">
          <p className="text-[10px] font-bold text-[#4f5660] dark:text-gray-400 uppercase tracking-wider">
            Не в сети — {offlineMembers.length}
          </p>
          {offlineMembers.map(member => renderMemberRow(member))}
        </div>
      )}

      {members.length === 0 && (
        <div className="text-xs text-gray-500 italic text-center py-4">Участников пока нет</div>
      )}
    </div>
  );
};

export default MemberList;