import React, { useState, useEffect } from 'react';
import { getServerMembers, Member } from '../../api/members';

interface MemberListProps {
  serverId: string | null;
}

const MemberList: React.FC<MemberListProps> = ({ serverId }) => {
  const [members, setMembers] = useState<Member[]>([]);

  useEffect(() => {
    if (!serverId) return;
  
    const fetchMembers = async () => {
      try {
        const data = await getServerMembers(serverId);
        
        // КРИТИЧНО: Проверяем, что пришел массив
        if (Array.isArray(data)) {
          setMembers(data);
        } else {
          // Если пришел объект (например, ошибка), логируем и ставим пустой массив
          console.error("Бэкенд вернул не массив участников:", data);
          setMembers([]); 
        }
      } catch (err) {
        console.error("Ошибка при загрузке участников:", err);
        setMembers([]); // Никогда не оставляй объект ошибки в состоянии для рендера
      }
    };
  
    fetchMembers();
  }, [serverId]);

  if (!serverId) return null;

  // Группируем участников на "В сети" и "Не в сети"
  const onlineMembers = members.filter(m => m.status === 'USER_STATUS_ONLINE');
  const offlineMembers = members.filter(m => m.status !== 'USER_STATUS_ONLINE');

  const renderMemberRow = (member: Member) => {
    // 1. Добавляем запасной вариант '?' на случай, если данных еще нет
    const displayName = member.nickname || member.username || 'Unknown';
    
    // 2. Безопасно берем первую букву
    const avatarLetter = displayName.charAt(0).toUpperCase() || '?';
    
    const isOnline = member.status === 'USER_STATUS_ONLINE';
  
    return (
      <div key={member.userId} className="flex items-center space-x-2 p-1.5 rounded hover:bg-[#35373c] cursor-pointer group">
        <div className="relative">
          <div className="w-8 h-8 bg-gray-600 rounded-full flex items-center justify-center font-bold text-xs uppercase text-gray-200">
            {/* Используем нашу безопасную переменную */}
            {avatarLetter}
          </div>
          <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-[#2b2d31] ${
            isOnline ? 'bg-[#23a55a]' : 'bg-gray-500'
          }`} />
        </div>
        <div className="text-sm truncate">
          <p className="font-medium text-gray-300 group-hover:text-white truncate">{displayName}</p>
          {member.nickname && <p className="text-[10px] text-gray-500 truncate">@{member.username}</p>}
        </div>
      </div>
    );
  };

  return (
    <div className="w-60 bg-[#2b2d31] border-l border-[#1e1f22] p-4 flex flex-col space-y-4 overflow-y-auto no-scrollbar hidden md:flex">
      {/* Секция "В сети" */}
      {onlineMembers.length > 0 && (
        <div className="space-y-1">
          <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">
            В сети — {onlineMembers.length}
          </p>
          {onlineMembers.map(member => (
            <div key={member.userId}> {/* <-- Ключ должен быть на самом верхнем элементе внутри map */}
              {renderMemberRow(member)}
            </div>
          ))}
        </div>
      )}

      {/* Секция "Не в сети" */}
      {offlineMembers.length > 0 && (
        <div className="space-y-1">
          <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">
            Не в сети — {offlineMembers.length}
          </p>
          {offlineMembers.map(renderMemberRow)}
        </div>
      )}

      {members.length === 0 && (
        <div className="text-xs text-gray-500 italic text-center py-4">Участников пока нет</div>
      )}
    </div>
  );
};

export default MemberList;