import { createContext, useContext, useState } from 'react';

interface MessagesSheetContextData {
  isOpen: boolean;
  selectedChatId: string | null;
  open: () => void;
  openChat: (chatId: string) => void;
  close: () => void;
  backToList: () => void;
}

export const MessagesSheetContext = createContext<MessagesSheetContextData>({} as MessagesSheetContextData);

export const MessagesSheetProvider = ({ children }: { children: React.ReactNode }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);

  const open = () => {
    setSelectedChatId(null);
    setIsOpen(true);
  };

  const openChat = (chatId: string) => {
    setSelectedChatId(chatId);
    setIsOpen(true);
  };

  const close = () => {
    setIsOpen(false);
  };

  const backToList = () => {
    setSelectedChatId(null);
  };

  return (
    <MessagesSheetContext.Provider value={{ isOpen, selectedChatId, open, openChat, close, backToList }}>
      {children}
    </MessagesSheetContext.Provider>
  );
};

export const useMessagesSheet = () => useContext(MessagesSheetContext);
