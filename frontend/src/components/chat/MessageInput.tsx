import React, { useState } from 'react';

interface MessageInputProps {
  onSend: (content: string) => void;
  placeholder: string;
}

const MessageInput: React.FC<MessageInputProps> = ({ onSend, placeholder }) => {
  const [text, setText] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;
    onSend(text);
    setText('');
  };

  return (
    <form onSubmit={handleSubmit} className="p-4 bg-white dark:bg-[#313338] transition-colors duration-200">
      <div className="bg-[#ebedef] dark:bg-[#383a40] rounded-lg px-4 py-2.5 flex items-center transition-colors">
        <button type="button" className="text-[#4f5660] dark:text-gray-400 hover:text-[#060607] dark:hover:text-white mr-3 text-xl">+</button>
        <input
          type="text"
          className="bg-transparent w-full text-[#060607] dark:text-gray-200 focus:outline-none text-[15px] placeholder-[#4f5660] dark:placeholder-gray-400"
          placeholder={placeholder}
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
      </div>
    </form>
  );
};

export default MessageInput;