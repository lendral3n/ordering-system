// src/components/Menu/Product360View.js
import React, { useRef } from "react";
import { Canvas, useLoader, useFrame } from "@react-three/fiber";
import { OrbitControls } from "@react-three/drei";
import * as THREE from "three";

const ImageSphere = ({ imageUrl }) => {
  const mesh = useRef();
  const texture = useLoader(THREE.TextureLoader, imageUrl);

  useFrame(() => {
    if (mesh.current) {
      mesh.current.rotation.y += 0.001;
    }
  });

  return (
    <mesh ref={mesh}>
      <sphereGeometry args={[5, 32, 32]} />
      <meshBasicMaterial map={texture} side={THREE.BackSide} />
    </mesh>
  );
};

const Product360View = ({ imageUrl }) => {
  return (
    <Canvas camera={{ position: [0, 0, 0.1], fov: 75 }}>
      <ImageSphere imageUrl={imageUrl} />
      <OrbitControls
        enableZoom={false}
        enablePan={false}
        autoRotate
        autoRotateSpeed={0.5}
      />
    </Canvas>
  );
};

export default Product360View;
