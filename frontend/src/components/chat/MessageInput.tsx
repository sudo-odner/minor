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
    <form onSubmit={handleSubmit} className="p-4 bg-[#313338]">
      <div className="bg-[#383a40] rounded-lg px-4 py-2.5 flex items-center">
        <button type="button" className="text-gray-400 hover:text-white mr-3 text-xl">+</button>
        <input
          type="text"
          className="bg-transparent w-full text-gray-200 focus:outline-none text-[15px]"
          placeholder={placeholder}
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
      </div>
    </form>
  );
};

export default MessageInput;