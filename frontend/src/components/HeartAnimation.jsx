import React from 'react';
import Lottie from 'lottie-react';
import heartAnimation from '../assets/heart-animation.json';

const HeartAnimation = ({ play, size = 100 }) => {
    return (
        <div className="heart-animation" style={{ width: size, height: size }}>
            <Lottie
                animationData={heartAnimation}
                loop={false}
                autoplay={play}
                style={{ width: '100%', height: '100%' }}
            />
        </div>
    );
};

export default HeartAnimation;