from setuptools import setup

name = 'offlinemsmtp'
version = '0.1'
release = '0.1.0'

setup(
    name=name,
    version=release,
    description='Offline msmtp wrapper',
    url='https://github.com/sumnerevans/offlinemsmtp',
    author='Sumner Evans',
    license='GPL3',
    packages=['offlinemsmtp'],
    install_requires=[
        'requests',
        'watchdog',
    ],
    zip_safe=False,

    # To provide executable scripts, use entry points in preference to the
    # "scripts" keyword. Entry points provide cross-platform support and
    # allow pip to create the appropriate form of executable for the target
    # platform.
    entry_points={
        'console_scripts': [
            'offlinemsmtp=offlinemsmtp.__main__:main',
        ],
    },
)
