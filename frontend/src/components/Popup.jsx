import React from "react";
import ReactDOM from "react-dom";

const Popup = ({ isOpen, title, message, onConfirm, onCancel, confirmText, cancelText, type }) => {
    if (!isOpen) return null;

    const getIconByType = () => {
        switch (type) {
            case 'delete':
                return (
                    <div className="popup-icon delete">
                        <svg width="24" height="24" viewBox="0 0 24 24">
                            <path fill="#e74c3c" d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                        </svg>
                    </div>
                );
            case 'warning':
                return (
                    <div className="popup-icon warning">
                        <svg width="24" height="24" viewBox="0 0 24 24">
                            <path fill="#f39c12" d="M12 2L1 21h22L12 2zm0 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                        </svg>
                    </div>
                );
            case 'auth':
                return (
                    <div className="popup-icon auth">
                        <svg width="24" height="24" viewBox="0 0 24 24">
                            <path fill="#3498db" d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/>
                        </svg>
                    </div>
                );
            default:
                return null;
        }
    };

    return ReactDOM.createPortal(
        <div className="popup-overlay">
            <div className="popup-container">
                {getIconByType()}
                <h3 className="popup-title">{title}</h3>
                <p className="popup-message">{message}</p>
                <div className="popup-actions">
                    <button className="popup-button cancel" onClick={onCancel}>
                        {cancelText || "Cancel"}
                    </button>
                    <button className="popup-button confirm" onClick={onConfirm}>
                        {confirmText || "Confirm"}
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
};

export default Popup;