from setuptools import setup, find_packages

setup(
    name="infragptapi",
    version="1.0.0",
    description="Python client for InfraGPT service",
    author="InfraGPT Team",
    packages=find_packages(),
    install_requires=[
        "grpcio>=1.50.0",
        "grpcio-tools>=1.50.0",
    ],
    python_requires=">=3.8",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
)